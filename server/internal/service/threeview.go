package service

import (
	"hunjie/internal/db"
	"hunjie/internal/dto"
	"hunjie/internal/model"
)

/*
 * 三观问答
 *
 * 设计要点（见 docs/08-权益体系重构方案.md 4.2② 与 6-3）：
 *
 *   - 答题免费，看答案需权益。若答题本身也要权益，就没人填了，
 *     而题库的全部价值恰恰在于「让所有人都填」——它不给平台增加任何
 *     数据采集成本，却能在无成本的前提下提供差异化。
 *   - 三观答案默认公开（是加分项，不敏感），与「原生家庭默认私有」相反。
 *     两条相反的策略都源于同一个判断：敏感度决定默认值。
 *   - 「看到对方的答案」这一动作也要权益，但对方是否答过题是公开信息，
 *     这样免费用户能看到「ta 答了 8 道题，解锁后可查看」的明确预期。
 *   - 匹配度只在双方都答过的题目上计算，不做「没答也算不匹配」的惩罚，
 *     否则会鼓励用户为凑数乱答。
 */

// QuestionIDs 题库的 ID 序列，用于把答案数组与题目对齐。
// 单独抽出来是因为题库 ID 是 1..N 连续的，但代码不应假设这一点。
func questionIDs() []int16 {
	ids := make([]int16, 0, len(model.ThreeViewQuestions))
	for _, q := range model.ThreeViewQuestions {
		ids = append(ids, q.ID)
	}
	return ids
}

// MyThreeViewQuestions 题库。前端答题页与展示页都用它渲染题目文本。
// 单独暴露是因为题库不随用户变化，可被前端缓存，不必每次都带答案一起拉。
func MyThreeViewQuestions() []model.ThreeViewQuestion {
	return model.ThreeViewQuestions
}

// MyThreeView 我的答题状态。返回题库和我的答案，前端渲染答题页/展示页都用它
func MyThreeView(me *model.User) (*dto.MyThreeViewResp, error) {
	answers, err := loadAnswers(me.ID)
	if err != nil {
		return nil, err
	}

	arr := make([]int8, 0, len(model.ThreeViewQuestions))
	answered := 0
	for _, q := range model.ThreeViewQuestions {
		if idx, ok := answers[q.ID]; ok {
			arr = append(arr, idx)
			answered++
		} else {
			arr = append(arr, -1)
		}
	}

	return &dto.MyThreeViewResp{
		Questions: model.ThreeViewQuestions,
		Answers:   arr,
		Answered:  answered,
		Total:     len(model.ThreeViewQuestions),
	}, nil
}

// SaveThreeView 全量覆盖我的答案。允许中途保存（未答用 -1 占位）。
//
// 幂等按 (user_id, q_id) 唯一索引实现：存在则更新，不存在则插入。
// 不做「先删后插」——那样中途保存会短暂清空历史，用户看到自己答案消失会很慌。
func SaveThreeView(me *model.User, req dto.SaveThreeViewReq) error {
	if len(req.Answers) != len(model.ThreeViewQuestions) {
		return Biz("答案数量与题目不匹配")
	}

	// 先校验一遍：任何一个下标越界就整体拒绝，避免写入半截脏数据
	for i, q := range model.ThreeViewQuestions {
		idx := req.Answers[i]
		if idx == -1 {
			continue // -1 表示未答，合法
		}
		if idx < 0 || int(idx) >= len(q.Options) {
			return Biz("第 %d 题的选项不正确", i+1)
		}
	}

	for i, q := range model.ThreeViewQuestions {
		idx := req.Answers[i]
		if idx == -1 {
			// 未答：删掉历史答案（用户主动清空某题）
			if err := db.DB.Where("user_id = ? AND q_id = ?", me.ID, q.ID).
				Delete(&model.ThreeViewAnswer{}).Error; err != nil {
				return err
			}
			continue
		}

		var exist model.ThreeViewAnswer
		err := db.DB.Where("user_id = ? AND q_id = ?", me.ID, q.ID).First(&exist).Error
		if err == nil {
			if exist.OptionIdx != idx {
				if err := db.DB.Model(&exist).Update("option_idx", idx).Error; err != nil {
					return err
				}
			}
			continue
		}
		if err := db.DB.Create(&model.ThreeViewAnswer{
			UserID: me.ID, QID: q.ID, OptionIdx: idx,
		}).Error; err != nil {
			// 并发重复提交：唯一索引冲突按幂等处理
			if isDuplicateKey(err) {
				continue
			}
			return err
		}
	}
	return nil
}

// ==================== 展示：他人资料页用 ====================

// buildThreeView 组装对方的三观展示数据。
//
// 参数 unlocked 是「我是否已解锁完整资料」。未解锁时不返回 Answers，
// 但 Answered 照常返回 —— 让免费用户知道「对方答过题，解锁后能看」，
// 而不是一片空白（空白会被误读成「对方没填」）。
func buildThreeView(me *model.User, targetID uint64, unlocked bool) *dto.ThreeViewInfo {
	peer, _ := loadAnswers(targetID)
	if len(peer) == 0 {
		return nil
	}

	info := &dto.ThreeViewInfo{
		Total:    len(model.ThreeViewQuestions),
		Answered: len(peer),
	}
	if !unlocked {
		return info
	}

	mine, _ := loadAnswers(me.ID)
	info.MyAnswers = make([]int8, 0, len(model.ThreeViewQuestions))
	info.Answers = make([]int8, 0, len(model.ThreeViewQuestions))

	comparable, agree := 0, 0
	for _, q := range model.ThreeViewQuestions {
		peerIdx, peerOK := peer[q.ID]
		if peerOK {
			info.Answers = append(info.Answers, peerIdx)
		} else {
			info.Answers = append(info.Answers, -1)
		}

		myIdx, myOK := mine[q.ID]
		if myOK {
			info.MyAnswers = append(info.MyAnswers, myIdx)
		} else {
			info.MyAnswers = append(info.MyAnswers, -1)
		}

		// 只统计「双方都答过」的题，未答不参与计算
		if peerOK && myOK {
			comparable++
			if peerIdx == myIdx {
				agree++
			}
		}
	}

	info.MyAnswered = len(mine) > 0
	if comparable > 0 {
		info.MatchRate = agree * 100 / comparable
	}
	return info
}

// loadAnswers 取某用户的答案，key 为题目 ID
func loadAnswers(userID uint64) (map[int16]int8, error) {
	var list []model.ThreeViewAnswer
	if err := db.DB.Where("user_id = ?", userID).Find(&list).Error; err != nil {
		return nil, err
	}
	out := make(map[int16]int8, len(list))
	for _, a := range list {
		out[a.QID] = a.OptionIdx
	}
	return out, nil
}
