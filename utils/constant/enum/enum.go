package enum

var (
	GenderEnum = map[int]string{
		0: "未记录",
		1: "男",
		2: "女",
	}
	RoleEnum = map[int]string{
		0: "未记录",
		1: "父亲",
		2: "母亲",
		3: "本人",
	}
	UserStatusEnum = map[int]string{
		0: "正常",
		1: "禁用",
	}
	RelationEnum = map[int]string{
		0: "未记录",
		1: "儿子",
		2: "女儿",
		3: "本人",
	}
	EducationEnum = map[int]string{
		0: "未记录",
		1: "初中",
		2: "高中",
		3: "专科",
		4: "本科",
		5: "硕士",
		6: "博士",
	}
	MaritalStatusEnum = map[int]string{
		0: "未记录",
		1: "未婚",
		2: "已婚",
		3: "离异",
		4: "丧偶",
	}
	MaritalPlanEnum = map[int]string{
		0: "未记录",
		1: "尽快",
		2: "一年内",
		3: "两年内",
		4: "时机成熟",
	}
	IsOnlyChildEnum = map[int]string{
		0: "未记录",
		1: "否",
		2: "是",
	}
	HouseEnum = map[int]string{
		0: "未记录",
		1: "暂未购房",
		2: "一套",
		3: "两套及以上",
	}
	CarEnum = map[int]string{
		0: "未记录",
		1: "暂未购车",
		2: "一辆",
		3: "两辆及以上",
	}
	YearIncomeEnum = map[int]string{
		0: "未记录",
		1: "10w以内",
		2: "10-20w",
		3: "20-30w",
		4: "30-50w",
		5: "50w以上",
	}
)
