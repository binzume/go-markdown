# Sample Markdown

テスト用ファイルです．

`go test` でのテスト時にhtmlが生成されます.

- [sample.md](sample.md)
- [sample.html](sample.html)

以下動作確認用.

## リスト

```md
- 新幹線
  - のぞみ
  - つばさ
  - あさま
- 特急
  - あずさ
  - しなの
```

- 新幹線
  - のぞみ
  - つばさ
  - あさま
- 特急
  - あずさ
  - しなの


1. aaa
1. bbb
1. ccc

- tasklist
- [ ] todo
- [x] done

## テーブル

```md
| 左寄せ | 中央 | 右寄せ |
|:------|:----:|------:|
| 1     |   2  |     3 |
| A     |   B  |     C |
```

| 左寄せ | 中央 | 右寄せ |
|:------|:----:|------:|
| 1     |   2  |     3 |
| A     |   B  |     C |

## コード

### Golang

```go
package main

import (
	"fmt"
	"github.com/binzume/go-calendar"
)

func main() {
	fmt.Println(calendar.NewCalendar().Markdown())
}
```

### C

```c
/*
	filter_source.pmのテスト
*/
int main()
{
	int test=123; // コメントテスト
	printf("Hello, World!\n");
	return 0;
}
```

### Ruby

```rb
# hello
class Hoge
  def hello
    puts "world!"
    puts 'class Hoge'
  end
end
```


# h1

## h2

### h3

#### h4

aaa `inline code` bbb.

hello.

|Table|test|
|1|2|
|3|4|

- item1
- item2 
- item3

- AutoLink https://www.google.co.jp
- Link [Google](https://www.google.co.jp)

