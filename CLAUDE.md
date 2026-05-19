# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o pydict2json .        # 编译
go run . "{'key': 'val'}"        # 直接运行

# 使用方式
echo "{'a': 1}" | ./pydict2json
pydict2json -f input.py -o output.json
pydict2json "{'a': [1, True], 'b': None}"
pydict2json -c "{'a': 1}"       # 紧凑输出
```

无测试文件、无 lint 配置。

## Architecture

单文件 Go 程序（`main.go`），将 Python dict 字面量转换为 JSON。核心是一个递归下降 parser（`Parser` struct），按首字符分派到子解析器：

- `parseDict` / `parseList` / `parseTuple` → 容器类型（tuple 序列化为 JSON array）
- `parseString` → 支持单引号、双引号、三引号，处理转义
- `parseNumber` → 整数 / 浮点（含科学计数法）
- `parseTrue` / `parseFalse` / `parseNone` → Python 布尔和 None

`orderedMap` 保持 Python dict 的插入顺序，自定义 `MarshalJSON` 实现。`jsonMarshalIndent` 先生成紧凑 JSON 再用标准库重新缩进。

输入来源优先级：命令行参数 > `-f` 文件 > stdin。
