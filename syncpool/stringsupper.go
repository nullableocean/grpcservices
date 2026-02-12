package syncpool

import (
	"sync"
	"unicode"
)

/*

## Оптимизация обработки строк с sync.Pool

### Описание задачи
В высоконагруженном сервисе частые аллокации буферов для преобразования строк создают нагрузку на GC.
Цель — реализовать оптимизированную функцию `ProcessString` с использованием `sync.Pool`, чтобы переиспользовать буферы `[]byte`.

### Требования
1. Функция `ProcessString(s string) string` преобразует строку в верхний регистр.
2. Использование `sync.Pool` для буферов `[]byte`.
3. Потокобезопасность, отсутствие утечек памяти.

*/

var (
	avgStringSize = 256
)

var bPool = &sync.Pool{
	New: func() interface{} { return make([]byte, 0, avgStringSize) },
}

func ProcessString(s string) string {
	buf := bPool.Get().([]byte)
	defer bPool.Put(buf)

	buf = buf[:0]
	for _, r := range s {
		buf = append(buf, byte(unicode.ToUpper(r)))
	}

	newStr := string(buf)
	return newStr
}
