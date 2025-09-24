<?php
namespace Chelbit\Exceltools\Data;

use JsonSerializable;

final class Sheet implements JsonSerializable
{
    private ?string $name;
    private ?int $index;

    /** @var Block[] */
    private array $blocks = [];

    private array $formulas = [];
    private array $options = [];
    private ?int $headerRow = null; // For import
    private ?int $startRow = null;  // For import

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

    public function addBlock(Block $block): self
    {
        $this->blocks[] = $block;
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
            'headerRow' => $this->headerRow,
            'startRow' => $this->startRow,
            'blocks' => $this->blocks,
            'formulas' => $this->formulas,
            'options' => $this->options,
        ];

        return array_filter($payload, fn($value) => $value !== null && $value !== []);
    }
}
