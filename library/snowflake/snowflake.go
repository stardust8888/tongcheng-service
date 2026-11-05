package snowflake

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

type Snowflake struct {
	mu            sync.Mutex
	epoch         int64 // 起始时间戳(毫秒)
	nodeID        int64 // 节点ID
	sequence      int64 // 序列号
	lastTimestamp int64 // 上次生成ID的时间戳
}

// 雪花ID各部分位数
const (
	timestampBits = 41 // 时间戳位数
	nodeIDBits    = 10 // 节点ID位数
	sequenceBits  = 12 // 序列号位数

	maxNodeID   = -1 ^ (-1 << nodeIDBits)   // 最大节点ID
	maxSequence = -1 ^ (-1 << sequenceBits) // 最大序列号

	timestampShift = sequenceBits + nodeIDBits // 时间戳左移位数
	nodeIDShift    = sequenceBits              // 节点ID左移位数

	// 默认起始时间戳 (2022-01-01 00:00:00 UTC)
	defaultEpoch = int64(1640995200000)
)

var (
	snowflake        *Snowflake
	once             sync.Once
	ErrInvalidNodeID = errors.New("node id must be between 0 and 1023")
	ErrTimeBackwards = errors.New("clock moved backwards")
	ErrInvalidEpoch  = errors.New("invalid epoch")
)

func InitSnowflake(nodeID int64) {
	if nodeID < 0 || nodeID > maxNodeID {
		nodeID = 0
	}

	once.Do(func() {
		snowflake = &Snowflake{
			epoch:         defaultEpoch,
			nodeID:        nodeID,
			sequence:      0,
			lastTimestamp: 0,
		}
	})
}

// SetEpoch 设置自定义起始时间戳
func SetEpoch(epoch int64) {
	if epoch <= 0 || epoch > time.Now().UnixNano()/1e6 {
		snowflake.epoch = defaultEpoch
	}

	if snowflake != nil {
		snowflake.mu.Lock()
		defer snowflake.mu.Unlock()
		snowflake.epoch = epoch
	}
}

func Generate() int64 {
	return snowflake.Generate()
}

func GenerateString() string {
	return strconv.Itoa(int(snowflake.Generate()))
}

func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 获取当前时间戳（毫秒）
	timestamp := time.Now().UnixNano() / 1e6

	// 如果当前时间小于上次生成ID的时间，说明时钟回拨了
	if timestamp < s.lastTimestamp {
		offset := s.lastTimestamp - timestamp + 1
		timestamp += offset
	}

	// 如果是同一毫秒内生成的ID，则序列号递增
	if timestamp == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		// 如果序列号溢出，则等待下一毫秒
		if s.sequence == 0 {
			timestamp = s.waitNextMillis(timestamp)
		}
	} else {
		// 不同毫秒内，序列号重置为0
		s.sequence = 0
	}

	// 更新上次生成ID的时间戳
	s.lastTimestamp = timestamp

	// 组装雪花ID
	id := ((timestamp - s.epoch) << timestampShift) |
		(s.nodeID << nodeIDShift) |
		s.sequence

	return id
}

// waitNextMillis 等待下一毫秒
func (s *Snowflake) waitNextMillis(timestamp int64) int64 {
	for timestamp <= s.lastTimestamp {
		timestamp = time.Now().UnixNano() / 1e6
	}
	return timestamp
}

// Parse 解析雪花ID，获取各部分信息
func Parse(id int64) (timestamp, nodeID, sequence int64) {
	if snowflake == nil {
		// 如果没有初始化，则使用默认epoch
		timestamp = ((id >> timestampShift) + defaultEpoch)
	} else {
		timestamp = ((id >> timestampShift) + snowflake.epoch)
	}

	nodeID = (id >> nodeIDShift) & maxNodeID
	sequence = id & maxSequence

	return
}

// GetTimestamp 从雪花ID中提取时间戳
func GetTimestamp(id int64) int64 {
	if snowflake == nil {
		return (id >> timestampShift) + defaultEpoch
	}
	return (id >> timestampShift) + snowflake.epoch
}

// GetNodeID 从雪花ID中提取节点ID
func GetNodeIDFrom(id int64) int64 {
	return (id >> nodeIDShift) & maxNodeID
}

// GetSequence 从雪花ID中提取序列号
func GetSequence(id int64) int64 {
	return id & maxSequence
}
