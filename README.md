# Микросервис для импорта/экспорта Excel и модуль для Bitrix

## Обзор

Этот проект представляет собой комплексное решение для импорта и экспорта сложных таблиц Excel. Он состоит из двух основных компонентов:

1.  **Go Микросервис (`excelms`)**: Высокопроизводительное, автономное приложение, которое может работать как HTTP-сервер или как инструмент командной строки (CLI). Оно обрабатывает основную логику парсинга и генерации файлов `.xlsx` на основе гибкой JSON-схемы.
2.  **PHP Модуль для Bitrix (`chelbit.exceltools`)**: Модуль для CMS Bitrix, который предоставляет удобную PHP-обертку для Go микросервиса, обеспечивая легкую интеграцию в проекты на Bitrix.

---

## 1. Go Микросервис (`excelms`)

Приложение на Go находится в директории `go/`.

### Возможности

-   **Два режима работы**: Запускается как HTTP API сервер или как CLI-приложение.
-   **Сложный экспорт**: Генерирует Excel-файлы из JSON-данных, определяя структуру через блоки, таблицы, колонки и строки. Поддерживает различную ориентацию данных (вертикальную/горизонтальную).
-   **Стили и формулы**: Позволяет применять пользовательские форматы ячеек, ширину колонок, закрепление областей, автофильтры и формулы при экспорте.
-   **Потоковый импорт**: Эффективно по использованию памяти парсит большие файлы Excel в поток NDJSON-объектов.
-   **Валидация данных**: Проверяет данные при импорте на соответствие определенной схеме, включая типы, обязательные поля, длину, регулярные выражения и др.

### Сборка и Установка

Для сборки бинарного файла из исходного кода:

```bash
# Перейдите в директорию go
cd go

# Соберите приложение
go build -o excelms .
```

Это создаст исполняемый файл `excelms` (или `excelms.exe` в Windows) в директории `go/`.

### Конфигурация

Приложение можно настроить с помощью переменных окружения или флагов командной строки.

#### Переменные окружения

-   `PORT`: Порт, на котором будет работать HTTP-сервер (по умолчанию: `8080`).
-   `AUTH_TOKEN`: Секретный Bearer-токен. Если он установлен, все эндпоинты `/api/*` потребуют заголовок `Authorization: Bearer <токен>`.

#### Флаги командной строки

-   `--serve` или `-s`: Запускает приложение в режиме сервера.
-   `--port <номер>`: Указывает порт сервера (имеет приоритет над переменной окружения `PORT`).

### Документация по API (Режим сервера)

#### `GET /health`

Простой эндпоинт для проверки работоспособности сервиса.

-   **Ответ `200 OK`**:
    ```json
    {"status":"ok"}
    ```

#### `POST /api/export`

Генерирует и возвращает файл `.xlsx`.

-   **Content-Type**: Может быть `application/json` или `multipart/form-data`.
-   **Тело запроса (`application/json`)**: JSON-объект, соответствующий структуре `ExportRequest`. Шаблон может быть включен в виде base64-строки в поле `templateBase64`.
-   **Тело запроса (`multipart/form-data`)**:
    -   `request`: Поле формы, содержащее JSON-объект для экспорта.
    -   `template`: Необязательное поле формы с файлом шаблона `.xlsx`.
-   **Ответ `200 OK`**: Сгенерированный файл `.xlsx` с `Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`.
-   **Ответ `4xx/5xx`**: JSON-объект с описанием ошибки.

**Пример тела запроса для экспорта:**

```json
{
  "sheet": "Месячный отчет",
  "columns": [
    {"name": "product", "header": "Продукт", "type": "string", "width": 30},
    {"name": "sales", "header": "Продажи", "type": "float", "customFormat": "#,##0.00"}
  ],
  "blocks": [
    {
      "startCell": "A2",
      "showHeaders": true,
      "data": [
        {"product": "Виджет A", "sales": 1500.50},
        {"product": "Виджет B", "sales": 2200.75}
      ]
    }
  ],
  "formulas": [
    {"column": "C", "startRow": 2, "endRow": 3, "expr": "=B{row}*1.2"}
  ],
  "options": {
    "freeze": "A2",
    "autofilter": "A1:B1"
  }
}
```

#### `POST /api/parse`

Парсит файл Excel и возвращает результаты в виде потока NDJSON.

-   **Content-Type**: `multipart/form-data`.
-   **Тело запроса**:
    -   `schema`: Поле формы, содержащее JSON-схему для импорта.
    -   `file`: Файл `.xlsx` для парсинга.
-   **Ответ `200 OK`**: Поток JSON-объектов, разделенных новой строкой (`application/x-ndjson`). Каждый объект представляет одну строку.

**Пример схемы для импорта:**

```json
{
  "sheet": "Sheet1",
  "headerRow": 1,
  "startRow": 2,
  "columns": [
    {
      "name": "name",
      "header": "Полное имя",
      "type": "string",
      "required": true
    },
    {
      "name": "age",
      "header": "Возраст",
      "type": "int",
      "min": 18
    }
  ]
}
```

**Пример ответа в формате NDJSON:**

```json
{"ok":true,"rowIndex":2,"data":{"name":"Иван Иванов","age":34}}
{"ok":false,"rowIndex":3,"data":{"name":"Мария Петрова","age":17},"errors":[{"field":"age","code":"min","msg":"Value must be at least 18"}]}
```

### Использование CLI

#### Экспорт

```bash
./excelms export --request <путь_до_request.json> --output <путь_для_export.xlsx> [--template <путь_до_шаблона.xlsx>]
```

-   `--request`: Путь к JSON-файлу, содержащему полное тело запроса (аналогично API).
-   `--output`: Путь для сохранения сгенерированного файла Excel.
-   `--template` (необязательно): Путь к файлу-шаблону Excel.

#### Импорт

```bash
./excelms import --schema <путь_до_схемы.json> --file <путь_до_файла_импорта.xlsx>
```

-   `--schema`: Путь к JSON-схеме импорта.
-   `--file`: Путь к файлу Excel для импорта.
-   Команда выводит результирующий NDJSON в стандартный вывод.

---

## 2. PHP Модуль для Bitrix (`chelbit.exceltools`)

PHP-модуль находится в директории `php/chelbit.exceltools/`.

### Установка

1.  Скопируйте директорию `php/chelbit.exceltools` в `<корень_вашего_сайта>/bitrix/modules/`.
2.  Перейдите в Админ-панель Bitrix -> Marketplace -> Установленные решения.
3.  Найдите в списке "Инструменты для Excel (Chelbit)" и нажмите "Установить".

### Настройка

После установки перейдите в Настройки -> Настройки модулей -> Инструменты для Excel.

-   **Режим работы**:
    -   `Сервис (HTTP API)`: Модуль будет взаимодействовать с Go-сервисом по HTTP.
    -   `Локальный бинарник (CLI)`: Модуль будет выполнять бинарный файл Go напрямую на сервере.
-   **URL Go-сервиса**: Базовый URL сервиса (например, `http://127.0.0.1:8080`).
-   **Токен авторизации**: Bearer-токен, если `AUTH_TOKEN` установлен на сервисе.
-   **Путь до бинарного файла**: Абсолютный путь до скомпилированного исполняемого файла `excelms`.

### Примеры использования

Модуль предоставляет простой и единый класс `Client` для всех операций.

#### Пример экспорта

```php
use Chelbit\Exceltools\Client;
use Chelbit\Exceltools\Schema;
use Chelbit\Exceltools\Column;
use Chelbit\Exceltools\Types;

// 1. Описываем колонки
$columns = [
    Column::byHeader('name', 'Полное имя', Types::STRING),
    Column::byHeader('email', 'Email', Types::STRING),
];

// 2. Создаем схему для листа
$schema = new Schema($columns, 'Список пользователей', 1, 2);

// 3. Готовим данные для блока `rows`
$rows = [
    ['name' => 'Иван Петров', 'email' => 'ivan@example.com'],
    ['name' => 'Анна Сидорова', 'email' => 'anna@example.com'],
];

// 4. Экспортируем файл
try {
    $client = new Client();
    $client->exportToFile(
        $_SERVER['DOCUMENT_ROOT'] . '/upload/users.xlsx',
        $rows,   // Передаем массив с данными
        $schema  // Передаем объект схемы
    );
    echo 'Экспорт успешно завершен!';
} catch (\Exception $e) {
    echo 'Ошибка: ' . $e->getMessage();
}
```

#### Пример импорта

```php
use Chelbit\Exceltools\Client;
use Chelbit\Exceltools\Schema;
use Chelbit\Exceltools\Column;
use Chelbit\Exceltools\Types;

// 1. Описываем схему для импортируемого файла
$schema = new Schema([
    Column::byHeader('name', 'Полное имя', Types::STRING)->setRequired(true),
    Column::byHeader('email', 'Email', Types::STRING),
]);

// 2. Импортируем и обрабатываем данные
try {
    $client = new Client();
    $filePath = $_SERVER['DOCUMENT_ROOT'] . '/upload/import_users.xlsx';

    foreach ($client->parse($filePath, $schema) as $row) {
        if ($row['ok']) {
            // Обрабатываем корректные данные
            print_r($row['data']);
        } else {
            // Обрабатываем ошибки для данной строки
            echo "Ошибка в строке {$row['rowIndex']}:\n";
            print_r($row['errors']);
        }
    }
} catch (\Exception $e) {
    echo 'Ошибка: ' . $e->getMessage();
}
```
