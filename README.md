# pgutils

Go-библиотека шаблонных функций для работы с СУБД PostgreSQL.

## Установка

```shell
go get -u github.com/zhikh23/pgutils
```

## Примеры использования

`Exec` предназначен для запросов, которые не возвращают результат.
Например, простой `INSERT INTO`.

```go
err := pgutils.Exec(ctx, db, `INSERT INTO users (id, name) VALUES ($1, $2)`, 1, "John")
if err != nil {
    ...
}
```

`Get` предназначен для запросов, которые могут вернуть не более одного результата.
Например, `SELECT` по первичному ключу.

```go
var name string
err := pgutils.Get(ctx, db, &name, `SELECT name FROM users WHERE id = $1`, 1)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        // not found
        ...
    }
    // another error
}
```

`Select` предназначен для запросов, которые могут вернуть произвольное количество результатов.

```go
var names []string
err := pgutils.Select(ctx, db, &names, `SELECT name FROM users`)
if err != nil {
    ...
}
```

`RunTx` оборачивает функцию в транзакцию.
В случае если функция возвращает ошибку, транзакция отменяется (`tx.Rollback`).
Если функция возвращает `nil`, транзакция автоматически применяется (`tx.Commit`).
Если завершение транзакции завершилось с ошибкой, объединяет ошибку закрытия транзакции и ошибку внутри транзакции.

```go
pgutils.RunTx(ctx, db,  func(tx *sqlx.Tx) error {
    err := pgutils.Exec(ctx, tx, `UPDATE users SET balance = 100 WHERE name = "Bob"`)
	if err != nil {
        return err
    }

    err = pgutils.Exec(ctx, tx, `UPDATE users SET balance = -100 WHERE name = "John"`)
	if err != nil {
        return err
    }
})
```
