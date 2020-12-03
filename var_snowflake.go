package var_snowflake

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

/*
生成的64bit整数, 组成部分由高位到低位如下:
(1) 16bit - 填充0
(2) 29bit - 时间戳, 单位秒, 表示从基线时间开始的秒数, 能支持17年
(3) 7bit  - node编号, 最大支持128个实例
(4) 12bit - seq序号, 最大支持同一秒内生成4096个ID

按照这个配置, 每秒最大ID生成数 = 128 * 4096 = 524288个.
生成的base64字符串最大长度为8.
*/

const (
	// URL安全的base64字符集(RFC 4648), 且经过乱序处理, 用于把int64转换成字符串
	encodeBase64Array = "Pl1i3Z9GTXgSuVB-KpxUbmER6FeA2v7o8zHYhcdajnM54rDfJkqI0wtCyLQ_OsWN"
	_64 = 64
	maxBase64StrLen = 8

	// 对64求模运算用的掩码
	mod64Mask int64 = (1 << 6) - 1

	// 除以64时位移的位数
	div64Shift = 6

	// node编号部分占7bit
	NodeBits uint8 = 7

	// seq序号部分占12bit
	StepBits uint8 = 12
)

// 用于把字符串转换成int64
var decodeBase64Map [128]byte

// 2020年1月1日0时0分0秒, 可作为基线时间使用
var Epoch20200101 time.Time = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// 预先把每个字符到64进制的映射保存下来
func init() {
	for i := 0; i < len(decodeBase64Map); i++ {
		decodeBase64Map[i] = 0xFF
	}

	for i := 0; i < len(encodeBase64Array); i++ {
		decodeBase64Map[encodeBase64Array[i]] = byte(i)
	}
}

// A snowflake id
type ID int64

// Snowflake generator node
type Node struct {
	mu    sync.Mutex
	epoch time.Time // 基线时间
	time  int64     // 距离基线时间隔了多少秒
	node  int64     // node编号
	step  int64     // seq序号

	nodeMax   int64 // 最大允许的node编号
	nodeMask  int64 // node编号部分的掩码
	stepMask  int64 // seq序号部分的掩码
	timeShift uint8 // 时间部分的offset
	nodeShift uint8 // node编号部分的offset
}

// 创建一个snowflake id生成器, epoch为基线时间, node为node编号
// 应当保证调用Generate()时, 机器当前时间大于epoch, 否则会出异常
func NewNode(epoch time.Time, node int64) (*Node, error) {
	n := &Node{
		epoch: epoch,
		node:  node,
	}
	n.nodeMax = -1 ^ (-1 << NodeBits) // 7 -> 127
	n.nodeMask = n.nodeMax << StepBits
	n.stepMask = -1 ^ (-1 << StepBits)
	n.timeShift = NodeBits + StepBits
	n.nodeShift = StepBits

	if n.node < 0 || n.node > n.nodeMax {
		return nil, fmt.Errorf("node number must be between 0 and %d", n.nodeMax)
	}

	return n, nil
}

// 生成一个ID
func (n *Node) Generate() ID {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Since(n.epoch).Nanoseconds() / 1e9 // 当前时间距离基线时间的秒数

	if now == n.time {
		n.step = (n.step + 1) & n.stepMask
		// 当1秒内生成的ID数量超出seq最大序号, 则忙等直到下一秒
		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Nanoseconds() / 1e9
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	id := ID((now)<<n.timeShift | (n.node << n.nodeShift) | (n.step))

	return id
}

// 以int64形式返回ID
func (id ID) Int64() int64 {
	return int64(id)
}

// 以2进制字符串形式返回ID
func (id ID) Base2() string {
	return strconv.FormatInt(int64(id), 2)
}

// 以base64形式返回ID
func (id ID) Base64() string {
	value := int64(id)
	if value < _64 {
		return string(encodeBase64Array[value])
	}

	b := make([]byte, 0, maxBase64StrLen)
	for value >= _64 {
		b = append(b, encodeBase64Array[value &mod64Mask])
		value = value >> div64Shift
	}
	b = append(b, encodeBase64Array[value])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

