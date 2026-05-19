# pydict2json

将 Python dict 字面量转换为 JSON 的命令行工具。

## 功能

- 解析 Python dict 语法，输出合法 JSON
- 支持的类型：dict、list、tuple、str（单引号/双引号/三引号）、int、float（含科学计数法）、True、False、None
- 保持 Python dict 的插入顺序
- 支持嵌套结构
- 支持尾逗号

## 安装

```bash
go build -o pydict2json .
```

## 使用

```bash
# 从 stdin 读取
echo "{'key': 'val', 'n': 42}" | pydict2json

# 从文件读取
pydict2json -f data.py -o data.json

# 命令行直接传入
pydict2json "{'a': [1, 2, True], 'b': None}"

# 紧凑输出
pydict2json -c "{'a': 1, 'b': 2}"
```

## 选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `-f <file>` | stdin | 输入文件 |
| `-o <file>` | stdout | 输出文件 |
| `-p` | true | 格式化输出（缩进） |
| `-c` | false | 紧凑输出（覆盖 `-p`） |
| `-h` | - | 显示帮助 |

## 示例

输入：

```python
{'name': 'Alice', 'scores': [95.5, 88, 92], 'active': True, 'note': None}
```

输出：

```json
{
  "name": "Alice",
  "scores": [95.5, 88, 92],
  "active": true,
  "note": null
}
```

## 类型映射

| Python | JSON |
|--------|------|
| dict | object |
| list | array |
| tuple | array |
| str | string |
| int | number |
| float | number |
| True | true |
| False | false |
| None | null |
