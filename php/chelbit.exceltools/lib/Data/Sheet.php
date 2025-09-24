<?php
namespace Chelbit\Exceltools\Data;

use JsonSerializable;

final class Sheet implements JsonSerializable
{
    private ?string $name;
    private ?int $index;

    /** @var Block[] */
    private array $blocks = [];

    /** @var Table[] */
    private array $tables = [];

    // Other sheet-level properties from Go struct can be added here if needed
    private array $formulas = [];
    private array $options = [];

    public function __construct(?string $name = null, ?int $index = null)
    {
        $this->name = $name;
        $this->index = $index;
    }

    public static function create(?string $name = null, ?int $index = null): self
    {
        return new self($name, $index);
    }

    public function getName(): ?string
    {
        return $this->name;
    }

    public function withBlocks(array $blocks): self
    {
        $this->blocks = $blocks;
        return $this;
    }

    public function addBlock(Block $block): self
    {
        $this->blocks[] = $block;
        return $this;
    }

    public function withTables(array $tables): self
    {
        $this->tables = $tables;
        return $this;
    }

    public function addTable(Table $table): self
    {
        $this->tables[] = $table;
        return $this;
    }

    public function forImport(int $headerRow, int $startRow): self
    {
        $this->headerRow = $headerRow;
        $this->startRow = $startRow;
        return $this;
    }

    public function jsonSerialize(): array
    {
        $payload = [
            'sheet' => $this->name,
            'sheetIndex' => $this->index,
            'blocks' => $this->blocks,
            'tables' => $this->tables,
            'formulas' => $this->formulas,
            'options' => $this->options,
        ];

        return array_filter($payload, fn($value) => $value !== null && $value !== []);
    }
}
