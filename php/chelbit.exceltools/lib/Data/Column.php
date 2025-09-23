<?php
namespace Chelbit\Exceltools\Data;

final class Column implements \JsonSerializable
{
    public string $name;
    public ?string $column = null;  // 'A','B',...
    public ?string $header = null;  // заголовок
    public string $type;

    // валидация
    public bool $required = false;
    public ?int $minLen = null;
    public ?int $maxLen = null;
    public ?string $regex = null;
    public array $enum = [];
    public ?float $min = null;
    public ?float $max = null;

    // даты
    public ?string $dateFormat = null;
    public ?string $timezone = null;

    // прочее
    public bool $trim = true;

    // экспорт
    public ?string $customFormat = null;
    public ?float $width = null;

    public function __construct(string $name, string $type){ $this->name=$name; $this->type=$type; }

    public static function byHeader(string $name, string $header, string $type): self { $c=new self($name,$type); $c->header=$header; return $c; }
    public static function byColumn(string $name, string $column, string $type): self { $c=new self($name,$type); $c->column=strtoupper($column); return $c; }

    /** Создать колонку из массива (как в JSON-схеме). */
    public static function fromArray(array $a): self
    {
        if (!isset($a['name'], $a['type'])) {
            throw new \InvalidArgumentException('Column.fromArray: required: name,type');
        }
        $c = new self((string)$a['name'], (string)$a['type']);
        $c->column = isset($a['column']) ? (string)$a['column'] : null;
        $c->header = isset($a['header']) ? (string)$a['header'] : null;
        $c->required = (bool)($a['required'] ?? false);
        $c->minLen = isset($a['minLen']) ? (int)$a['minLen'] : null;
        $c->maxLen = isset($a['maxLen']) ? (int)$a['maxLen'] : null;
        $c->regex  = isset($a['regex']) ? (string)$a['regex'] : null;
        $c->enum   = isset($a['enum']) && is_array($a['enum']) ? array_values($a['enum']) : [];
        $c->min    = isset($a['min']) ? (float)$a['min'] : null;
        $c->max    = isset($a['max']) ? (float)$a['max'] : null;
        $c->dateFormat = isset($a['dateFormat']) ? (string)$a['dateFormat'] : null;
        $c->timezone   = isset($a['timezone']) ? (string)$a['timezone'] : null;
        $c->trim       = isset($a['trim']) ? (bool)$a['trim'] : true;
        $c->customFormat = isset($a['customFormat']) ? (string)$a['customFormat'] : null;
        $c->width = isset($a['width']) ? (float)$a['width'] : null;
        return $c;
    }

    public function jsonSerialize(): mixed
    {
        return array_filter(get_object_vars($this), fn($v) => $v !== null);
    }
}
