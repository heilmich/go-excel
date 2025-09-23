# Микросервис для импорта/экспорта Excel и модуль для Bitrix

## Обзор

Этот проект представляет собой комплексное решение для импорта и экспорта сложных таблиц Excel. Он состоит из двух основных компонентов:

1.  **Go Микросервис (`excelms`)**: Высокопроизводительное, автономное приложение, которое может работать как HTTP-сервер или как инструмент командной строки (CLI). Оно обрабатывает основную логику парсинга и генерации файлов `.xlsx` на основе гибкой JSON-схемы.
2.  **PHP Модуль для Bitrix (`chelbit.exceltools`)**: Модуль для CMS Bitrix, который предоставляет удобную PHP-обертку для Go микросервиса, обеспечивая легкую интеграцию в проекты на Bitrix.

---

## 1. Go Микросервис (`excelms`)

... (документация по Go опущена для краткости) ...

### API: `POST /api/parse` (Импорт)

Парсит файл Excel в соответствии со схемой и возвращает извлеченные данные в виде единого JSON-объекта.

-   **Content-Type**: `multipart/form-data`.
-   **Тело запроса**:
    -   `schema`: Поле формы, содержащее JSON-схему для импорта. Схема должна определять Листы и Блоки/Таблицы, из которых нужно извлечь данные. В каждом блоке можно указать опциональный строковый `id`.
    -   `file`: Файл `.xlsx` для парсинга.
-   **Ответ `200 OK`**: Единый JSON-объект, сгруппированный по листам и идентификаторам блоков.

**Пример ответа импорта:**

```json
{
  "data": {
    "ИменаСотрудников": {
      "main_users_block": [
        { "name": "Иван", "surname": "Петров" },
        { "name": "Анна", "surname": "Сидорова" }
      ]
    }
  },
  "errors": []
}
```

---

## 2. PHP Модуль для Bitrix (`chelbit.exceltools`)

... (установка и настройка опущены для краткости) ...

### Примеры использования

#### Пример экспорта (с Блоками и Таблицами)

```php
use Chelbit\Exceltools\Client;
use Chelbit\Exceltools\Data\Schema;
use Chelbit\Exceltools\Data\Sheet;
use Chelbit\Exceltools\Data\Table;
use Chelbit\Exceltools\Data\Column;
use Chelbit\Exceltools\Enums\Types;

$schema = Schema::create([
    Sheet::create('Пользователи')->addBlock(
        Table::create('A1', [
            (new Column('name', Types::STRING))->setHeader('Имя'),
        ], [['name' => 'Иван']])
        ->withId('users_table') // Задаем ID для блока
    )
]);

try {
    $client = new Client();
    $client->exportToFile('report.xlsx', $schema);
} catch (\Exception $e) { /* ... */ }
```

#### Пример импорта (по Блокам)

```php
use Chelbit\Exceltools\Client;
use Chelbit\Exceltools\Data\Schema;
use Chelbit\Exceltools\Data\Sheet;
use Chelbit\Exceltools\Data\Block;
use Chelbit\Exceltools\Data\Column;
use Chelbit\Exceltools\Enums\Types;

// 1. Описываем, откуда и какие данные мы хотим прочитать
$schema = Schema::create([
    Sheet::create('Пользователи')->addBlock(
        Block::create('A1', [
            (new Column('name', Types::STRING)),
            (new Column('email', Types::STRING)),
        ])
        ->withId('main_users_block') // Указываем ID, чтобы легко найти данные в ответе
    )
]);

// 2. Импортируем и обрабатываем данные
try {
    $client = new Client();
    $filePath = $_SERVER['DOCUMENT_ROOT'] . '/upload/users_to_import.xlsx';

    $result = $client->parse($filePath, $schema);

    // Получаем данные из конкретного блока по его ID
    $userData = $result['data']['Пользователи']['main_users_block'] ?? [];
    foreach ($userData as $user) {
        print_r($user);
    }

} catch (\Exception $e) {
    echo 'Ошибка: ' . $e->getMessage();
}
```
