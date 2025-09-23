<?php
namespace Chelbit\Exceltools;

final class Schema implements \JsonSerializable
{
    public ?string $sheet = null;
    public int $headerRow = 1;
    public int $startRow = 2;
    /** @var Column[] */
    public array $columns = [];

    public function __construct(array $columns, ?string $sheet=null, int $headerRow=1, int $startRow=2)
    {
        $this->columns = $columns; $this->sheet = $sheet; $this->headerRow = $headerRow; $this->startRow = $startRow;
    }

    /** Универсальный конструктор: принимает Schema|array|json-string. */
    public static function fromMixed($v): self
    {
        if ($v instanceof self) { return $v; }
        if (is_string($v)) { return self::fromJson($v); }
        if (is_array($v)) { return self::fromArray($v); }
        throw new \InvalidArgumentException('Schema.fromMixed: ожидается Schema|array|json-string');
    }

    /** Создать из JSON-строки. */
    public static function fromJson(string $json): self
    {
        $a = json_decode($json, true);
        if (!is_array($a)) {
            throw new \InvalidArgumentException('Schema.fromJson: не удалось разобрать JSON');
        }
        return self::fromArray($a);
    }

    /** Создать из ассоциативного массива (как в JSON). */
    public static function fromArray(array $a): self
    {
        $cols = [];
        if (!isset($a['columns']) || !is_array($a['columns']) || count($a['columns']) === 0) {
            throw new \InvalidArgumentException('Schema.fromArray: columns пуст');
        }
        foreach ($a['columns'] as $c) {
            if ($c instanceof Column) { $cols[] = $c; }
            elseif (is_array($c)) { $cols[] = Column::fromArray($c); }
            else { throw new \InvalidArgumentException('Schema.fromArray: неподдерживаемая колонка'); }
        }
        $sheet = isset($a['sheet']) ? (string)$a['sheet'] : null;
        $headerRow = isset($a['headerRow']) ? (int)$a['headerRow'] : 1;
        $startRow  = isset($a['startRow']) ? (int)$a['startRow'] : 2;
        return new self($cols, $sheet, $headerRow, $startRow);
    }

    public function jsonSerialize(): mixed
    {
        return [
            'sheet'=>$this->sheet, 'headerRow'=>$this->headerRow, 'startRow'=>$this->startRow,
            'columns'=>array_map(fn(Column $c)=>$c->jsonSerialize(), $this->columns),
        ];
    }
}
