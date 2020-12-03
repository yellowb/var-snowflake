package snowflake

// URL安全的base64字符集(RFC 4648), 且经过乱序处理, 用于把int64转换成字符串
const encodeBase64Array = "Pl1i3Z9GTXgSuVB-KpxUbmER6FeA2v7o8zHYhcdajnM54rDfJkqI0wtCyLQ_OsWN"

// 用于把字符串转换成int64
var decodeBase64Map [128]byte