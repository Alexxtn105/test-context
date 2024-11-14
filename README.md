# Тестовый проект с использованием контекстов

Пакет `context` определяет тип `Context`, который позволяет управлять дедлайнами, сигналами отмены и другими значениями области действия запросов между границами API и процессами.

Это объект который в первую очередь предназначен для того, чтобы иметь возможность отменить извне потенциально долгую операцию.
Кроме этого с помощью контекста можно передавать информацию между функциями и методами.

Отмена производится несколькими способами:
- По явному сигнал отмены - context.WithCancel()
- По истечению промежутка времени - context.WithTimeout()
- По достижению определенной временной отметки (по "дедлайну") - context.WithDeadline()

# Когда использовать контекст:
- Метод ходит куда-то по сети
- Горутина может долго висеть

# Советы
- Передавать всегда первым аргументом
- Передавать только в функции и методы, не хранить его в состояниях (внутри структуры), поскольку они спроектированы для одноразового использования и только для чтения
- Использовать `context.WithValue` только в исключительных случаях
- `context.Background` должен использоваться как родитель
- `context.TODO` используется, когда нет уверенности, какой он будет. Это всего-лишь заглушка, не предоставляющая средств контроля
- Не забывать про функции отмены контекста
- Передавать только контекст, без функции отмены. Контроль за завершением контекста должен быть на вызывающей стороне, иначе приложение может стать очень запутанным