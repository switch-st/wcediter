package models

// CharacterData 角色属性数据结构
type CharacterData struct {
	CurrentExp   int32 // 当前经验值
	NextLevelExp int32 // 升级经验值
	CurrentHP    int32 // 当前生命值
	CurrentMP    int32 // 当前内力值
	MaxHP        int32 // 最大生命值
	MaxMP        int32 // 最大内力值
	Strength     int16 // 力量
	Reaction     int16 // 反应
	Constitution int16 // 体质
	Speed        int16 // 速度
	Attack       int16 // 攻击
	Defense      int16 // 防御
	Luck         int16 // 运气
	Level        int16 // 等级
}

// RawByteData 保存原始字节的结构
type RawByteData struct {
	CurrentExp   []byte
	NextLevelExp []byte
	CurrentHP    []byte
	CurrentMP    []byte
	MaxHP        []byte
	MaxMP        []byte
	Strength     []byte
	Reaction     []byte
	Constitution []byte
	Speed        []byte
	Attack       []byte
	Defense      []byte
	Luck         []byte
	Level        []byte
}

// CharacterInfo 角色信息结构体
type CharacterInfo struct {
	Name     string
	Data     CharacterData
	RawBytes RawByteData
	Position int64 // 记录角色数据在文件中的起始位置
}

// MoneyInfo 银两信息结构体
type MoneyInfo struct {
	Value    int32
	RawBytes []byte
	Position int64
}

// ProgressInfo 进度信息结构体
type ProgressInfo struct {
	ProgressID   int    // 进度编号
	LocationID   int    // 位置编号
	LocationName string // 位置名称
}
