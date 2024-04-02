package raft

import "testing"

func TestCampaignScene(t *testing.T) {
	// 当超时时间到，发起选举
	// 发起者：
	// 1. 首先会转变状态， 将自己变为候选人
	// 2. 递增当前的任期号（currentTerm）
	// 3. 发起投票请求
	// 接收者：
	// 1. 判断如果自己的任期号小于投票请求的任期号，则转变为跟随者, 并回复同意
	// 2. 如果任期号相同，则首先判断当前是否已经给该任期投过票， 如果未投过，检查日志是否和自己一样新，如果一样投票赞成。 如果已经投过票， 检查请求是否
	//    是否来自于投过票的节点，如果是投票同意，其他情况拒绝
	// 3. 如果当前任期号大于投票请求的任期号，则拒绝投票
}
