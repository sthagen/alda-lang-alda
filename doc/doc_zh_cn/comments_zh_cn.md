# 注释(Comments)

*本页面翻译自[Comments](../comments.md)*

您可以通过在一行的左侧添加`#`符号来在Alda中添加一行**注释**(将该行*注释掉*)

```alda
# 这是一行注释
piano: c d e f
```

Alda会忽略掉以`#`号开头的行*(译者注: 如果您不想某些代码被演奏 但是又想在文档中保留那些代码 可以将它们做成注释 这样它们就不会被执行了)*

```alda
# trumpet: c c c c   <- 您不会听到这个
piano: c d e f     # <- 您会听到这个
```

