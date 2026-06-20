# internal/dao/errors.go

定义 DAO 层的统一错误。

## 变量

- `ErrNotFound`：按条件未查询到记录。各实现（如 [`entdao`](./entdao/db.md)）需将底层「记录不存在」错误（如 `ent.IsNotFound`）统一转换为该错误，使业务层无需感知具体存储后端的错误类型，可用 `errors.Is(err, dao.ErrNotFound)` 判断。
