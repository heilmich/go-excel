<?php
namespace Chelbit\Exceltools\Data;

use JsonSerializable;

/**
 * Represents the entire file schema, containing one or more sheets.
 */
final class Schema implements JsonSerializable
{
    /** @var Sheet[] */
    private array $sheets = [];

    public function __construct(array $sheets = [])
    {
        $this->sheets = $sheets;
    }

    public static function create(array $sheets = []): self
    {
        return new self($sheets);
    }

    public function addSheet(Sheet $sheet): self
    {
        $this->sheets[] = $sheet;
        return $this;
    }

    public function jsonSerialize(): array
    {
        // The Go service will be updated to expect this nested structure.
        return [
            'sheets' => $this->sheets,
        ];
    }
}
