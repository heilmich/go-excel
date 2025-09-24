<?php
namespace Chelbit\Exceltools\Data;

use JsonSerializable;

class Block implements JsonSerializable
{
    protected ?string $id = null;
    protected ?string $startCell = null;
    protected ?int $startRow = null;
    protected ?string $startCol = null;
    protected string $orientation = 'vertical';
    protected bool $showHeaders = false; // Default for a Block is false

    protected array $columns = [];
    protected array $data = [];

    public function __construct(string $startCell, array $columns, array $data)
    {
        $this->startCell = $startCell;
        $this->columns = $columns;
        $this->data = $data;
    }

    public static function create(string $startCell, array $columns, array $data): self
    {
        return new static($startCell, $columns, $data); // Use `static` for late static binding
    }

    public function withId(string $id): self
    {
        $this->id = $id;
        return $this;
    }

    public function setOrientation(string $orientation): self
    {
        $this->orientation = $orientation;
        return $this;
    }

    public function withHeaders(bool $show = true): self
    {
        $this->showHeaders = $show;
        return $this;
    }

    public function setStartRowCol(int $row, string $col): self
    {
        $this->startCell = null;
        $this->startRow = $row;
        $this->startCol = $col;
        return $this;
    }

    public function jsonSerialize(): array
    {
        $payload = [
            'id' => $this->id,
            'startCell' => $this->startCell,
            'startRow' => $this->startRow,
            'startCol' => $this->startCol,
            'orientation' => $this->orientation,
            'showHeaders' => $this->showHeaders,
            'columns' => $this->columns, // Now Block also has columns
            'data' => $this->data,
        ];

        return array_filter($payload, fn($value) => $value !== null);
    }
}
