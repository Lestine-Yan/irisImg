package dao

import "errors"

// ErrNotFound 表示按条件未查询到记录。
// 各实现需将底层「记录不存在」错误统一转换为该错误，
// 以便业务层无需感知具体存储后端的错误类型。
var ErrNotFound = errors.New("record not found")
