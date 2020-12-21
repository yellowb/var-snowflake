package var_snowflake

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

/*
生成的64bit整数, 组成部分由高位到低位如下:
(1) 11bit - 填充0
(2) 8bit  - 随机数
(3) 29bit - 时间戳, 单位秒, 表示从基线时间开始的秒数, 能支持17年
(4) 6bit  - node编号, 最大支持64个实例
(5) 10bit - seq序号, 最大支持同一秒内生成1024个ID

按照这个配置, 每秒最大ID生成数 = 64 * 1024 = 64K个.
由于保证了至少高11bit都为0, 所以生成的base64字符串最大长度为9. Log64(2 ^ 54 - 1) = Log64(18014398509481983) < 9
*/

const (
	// URL安全的base64字符集(RFC 4648), 且经过乱序处理, 用于把int64转换成字符串
	encodeBase64Array = "Pl1i3Z9GTXgSuVB-KpxUbmER6FeA2v7o8zHYhcdajnM54rDfJkqI0wtCyLQ_OsWN"
	_64               = 64
	maxBase64StrLen   = 9 // 允许生成的Base64字符串长度=9

	// 对64求模运算用的掩码
	mod64Mask int64 = (1 << 6) - 1

	// 除以64时位移的位数
	div64Shift = 6

	// node编号部分占6bit
	NodeBits uint8 = 6

	// seq序号部分占10bit
	StepBits uint8 = 10

	// 时间戳部分占29bit
	TimeBits uint8 = 29

	// 随机数部分占8bit
	RandomBits uint8 = 8
)

// 用于对调int64里的某些bit
var swapBitPairs = [][]int64{
	[]int64{24, 53},
	[]int64{27, 52},
	[]int64{30, 51},
	[]int64{33, 50},
	[]int64{36, 49},
	[]int64{39, 48},
	[]int64{42, 47},
	[]int64{45, 46},
}

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

	stepMask    int64 // seq序号部分的掩码
	nodeShift   uint8 // node编号部分的offset
	nodeMax     int64 // 最大允许的node编号
	nodeMask    int64 // node编号部分的掩码
	timeShift   uint8 // 时间部分的offset
	timeMax     int64 // 最大允许的时间戳
	timeMask    int64 // 时间部分的掩码
	randomShift uint8 // 随机数部分的offset
}

// 创建一个snowflake id生成器, epoch为基线时间, node为node编号
// 应当保证调用Generate()时, 机器当前时间大于epoch, 否则会出异常
func NewNode(epoch time.Time, node int64) (*Node, error) {
	n := &Node{
		epoch: epoch,
		node:  node,
	}
	n.stepMask = -1 ^ (-1 << StepBits)
	n.nodeShift = StepBits
	n.nodeMax = -1 ^ (-1 << NodeBits) // 6 -> 63
	n.nodeMask = n.nodeMax << StepBits
	n.timeShift = NodeBits + StepBits
	n.timeMax = -1 ^ (-1 << TimeBits)
	n.timeMask = n.timeMax << n.timeShift
	n.randomShift = n.timeShift + TimeBits

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

	randomNum := rand.Int63n(256) // 有8个bit可以表示随机数

	id := ID((randomNum)<<n.randomShift | (now)<<n.timeShift | (n.node << n.nodeShift) | (n.step))
	id = n.shuffleBits(id) // 把random部分的几个bit跟其它部分对调, 造成一种随机的效果

	return id
}

// 把生成的int64里的某些bit位置对调, 产生一定的随机效果
func (n *Node) shuffleBits(src ID) ID {
	var newId ID = src
	for _, swapPair := range swapBitPairs {
		newId = n.swapTwoBitsInInt64(newId, swapPair[0], swapPair[1])
	}

	return newId
}

// 交换int64里位置为p1和p2的两个bit, 位置从0开始计数
func (n *Node) swapTwoBitsInInt64(src ID, p1, p2 int64) ID {
	src = src ^ (1 << p1)
	src = src ^ (1 << p2)
	return src
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
		b = append(b, encodeBase64Array[value&mod64Mask])
		value = value >> div64Shift
	}
	b = append(b, encodeBase64Array[value])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}
