# var-snowflake
瘦身的雪花算法

### 简介
生成的64bit整数, 组成部分由高位到低位如下:

1. 17bit - 填充0
2. 29bit - 时间戳, 单位秒, 表示从基线时间开始的秒数, 能支持17年
3. 6bit  - node编号, 最大支持64个实例
4. 12bit - seq序号, 最大支持同一秒内生成4096个ID

按照这个配置, 每秒最大ID生成数 = 64 * 4096 = 256K个.

由于保证了至少高17bit都为0, 所以生成的base64字符串最大长度为8. Log64(2 ^ 48 - 1) = Log64(281474976710655) < 8

### 使用方法:
见`var_snowflake_test.go`文件。

### 测试结果:
以下测试结果都以2020年1月1日0时0分0秒作为基线时间
#### 1.电脑时间为2020年12月4日
    === RUN   TestNode_Generate
    lfbjClPP = 1101111010100101000110111000001000000000000 = 7650022264832
    lfbjClPl = 1101111010100101000110111000001000000000001 = 7650022264833
    lfbjClP1 = 1101111010100101000110111000001000000000010 = 7650022264834
    lfbjClPi = 1101111010100101000110111000001000000000011 = 7650022264835
    lfbjClP3 = 1101111010100101000110111000001000000000100 = 7650022264836
    -----------------
    lfbjylPP = 1101111010100101000111000000001000000000000 = 7650022526976
    lfbjylPl = 1101111010100101000111000000001000000000001 = 7650022526977
    lfbjylP1 = 1101111010100101000111000000001000000000010 = 7650022526978
    lfbjylPi = 1101111010100101000111000000001000000000011 = 7650022526979
    lfbjylP3 = 1101111010100101000111000000001000000000100 = 7650022526980
    -----------------
    lfbjLlPP = 1101111010100101000111001000001000000000000 = 7650022789120
    lfbjLlPl = 1101111010100101000111001000001000000000001 = 7650022789121
    lfbjLlP1 = 1101111010100101000111001000001000000000010 = 7650022789122
    lfbjLlPi = 1101111010100101000111001000001000000000011 = 7650022789123
    lfbjLlP3 = 1101111010100101000111001000001000000000100 = 7650022789124
    -----------------
    --- PASS: TestNode_Generate (3.00s)

#### 2.把电脑时钟调到2035年12月4日
    === RUN   TestNode_Generate
    vO0VLlPP = 11101111100110100001101111001000001000000000000 = 131722585051136
    vO0VLlPl = 11101111100110100001101111001000001000000000001 = 131722585051137
    vO0VLlP1 = 11101111100110100001101111001000001000000000010 = 131722585051138
    vO0VLlPi = 11101111100110100001101111001000001000000000011 = 131722585051139
    vO0VLlP3 = 11101111100110100001101111001000001000000000100 = 131722585051140
    -----------------
    vO0VQlPP = 11101111100110100001101111010000001000000000000 = 131722585313280
    vO0VQlPl = 11101111100110100001101111010000001000000000001 = 131722585313281
    vO0VQlP1 = 11101111100110100001101111010000001000000000010 = 131722585313282
    vO0VQlPi = 11101111100110100001101111010000001000000000011 = 131722585313283
    vO0VQlP3 = 11101111100110100001101111010000001000000000100 = 131722585313284
    -----------------
    vO0V_lPP = 11101111100110100001101111011000001000000000000 = 131722585575424
    vO0V_lPl = 11101111100110100001101111011000001000000000001 = 131722585575425
    vO0V_lP1 = 11101111100110100001101111011000001000000000010 = 131722585575426
    vO0V_lPi = 11101111100110100001101111011000001000000000011 = 131722585575427
    vO0V_lP3 = 11101111100110100001101111011000001000000000100 = 131722585575428
    -----------------
    --- PASS: TestNode_Generate (3.00s)
