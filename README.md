# Микросервис для импорта/экспорта Excel и модуль для Bitrix

## Обзор

Этот проект представляет собой комплексное решение для импорта и экспорта сложных таблиц Excel. Он состоит из двух основных компонентов:

1.  **Go Микросервис (`excelms`)**: Высокопроизводительное, автономное приложение, которое может работать как HTTP-сервер или как инструмент командной строки (CLI). Оно обрабатывает основную логику парсинга и генерации файлов `.xlsx` на основе гибкой JSON-схемы.
2.  **PHP Модуль для Bitrix (`chelbit.exceltools`)**: Модуль для CMS Bitrix, который предоставляет удобную PHP-обертку для Go микросервиса, обеспечивая легкую интеграцию в проекты на Bitrix.

---

## 1. Go Микросервис (`excelms`)

... (документация по Go опущена для краткости) ...

---

## 2. PHP Модуль для Bitrix (`chelbit.exceltools`)

... (установка и настройка опущены для краткости) ...

### Примеры использования

#### Пример экспорта (с Блоками, Таблицами и Шаблоном)

```php
use Chelbit\Exceltools\Client;
use Chelbit\Exceltools\Data\Schema;
use Chelbit\Exceltools\Data\Sheet;
use Chelbit\Exceltools\Data\Block;
use Chelbit\Exceltools\Data\Table;
use Chelbit\Exceltools\Data\Column;
use Chelbit\Exceltools\Enums\Types;

// 1. Готовим данные
$users = [
    ['name' => 'Иван Петров', 'email' => 'ivan@example.com'],
];
$stats = [
    ['month' => 'Январь', 'revenue' => 50000],
];

// 2. Описываем структуру через объекты
$schema = Schema::create([
    Sheet::create('Пользователи и Статистика')
        ->addBlock(
            // Table - это Блок, который по умолчанию выводит заголовки.
            Table::create('A1', [
                (new Column('name', Types::STRING))->setHeader('Имя'),
                (new Column('email', Types::STRING))->setHeader('Email')->setWidth(30),
            ], $users)->withId('users_table')
        )
        ->addBlock(
            // Block - по умолчанию НЕ выводит заголовки.
            Block::create('A10', [
                (new Column('month', Types::STRING))->setHeader('Месяц'),
                (new Column('revenue', Types::FLOAT))->setHeader('Выручка'),
            ], $stats)->withId('stats_block')->withHeaders(true) // но их можно включить
        )
]);

// 3. Экспортируем файл, используя существующий файл как шаблон
try {
    $client = new Client();
    // Указываем путь к файлу-шаблону. Стили и другие листы из него сохранятся.
    $templatePath = $_SERVER['DOCUMENT_ROOT'] . '/upload/templates/report_template.xlsx';

    $client->exportToFile(
        $_SERVER['DOCUMENT_ROOT'] . '/upload/final_report.xlsx',
        $schema,
        $templatePath // <--- Передача шаблона
    );
    echo 'Экспорт успешно завершен!';
} catch (\Exception $e) {
    echo 'Ошибка: ' . $e->getMessage();
}
```

#### Пример импорта (по Блокам)

```php
use Chelbit\Exceltools\Client;
use Chelbit\Exceltools\Data\Schema;
use Chelbit\Exceltools\Data\Sheet;
use Chelbit\Exceltools\Data\Block;
use Chelbit\Exceltools\Data\Column;
use Chelbit\Exceltools\Enums\Types;

// Описываем, откуда и какие данные мы хотим прочитать
$schema = Schema::create([
    Sheet::create('Пользователи')->addBlock(
        Block::create('A1', [
            (new Column('name', Types::STRING)),
            (new Column('email', Types::STRING)),
        ])->withId('main_users_block')
    )
]);

// ... (логика импорта)
```
