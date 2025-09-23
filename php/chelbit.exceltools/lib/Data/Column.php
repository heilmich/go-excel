<?php
namespace Chelbit\Exceltools\Data;

use JsonSerializable;

final class Column implements JsonSerializable
{
    public string $name;
    public string $type;
    public ?string $header;
    public ?string $column;

    public ?string $customFormat;
    public ?float $width;

    public bool $required;
    public bool $trim;
    public ?int $minLen;
    public ?int $maxLen;
    public ?string $regex;
    public array $enum;
    public ?float $min;
    public ?float $max;

    public ?string $dateFormat;
    public ?string $timezone;

    public function __construct(
        string $name,
        string $type,
        ?string $header = null,
        bool $required = false,
        // Add all other properties as optional constructor args if needed
    ) {
        $this->name = $name;
        $this->type = $type;
        $this->header = $header;
        $this->required = $required;

        // Set defaults for non-nullable properties
        $this->trim = true;
        $this->enum = [];
    }

    public static function create(string $name, string $type, ?string $header = null, bool $required = false): self
    {
        return new self($name, $type, $header, $required);
    }

    // Fluent setters for optional properties
    public function setHeader(string $header): self { $this->header = $header; return $this; }
    public function setColumn(string $column): self { $this->column = strtoupper($column); return $this; }
    public function setRequired(bool $required = true): self { $this->required = $required; return $this; }
    public function setWidth(float $width): self { $this->width = $width; return $this; }
    public function setCustomFormat(string $format): self { $this->customFormat = $format; return $this; }
    // ... etc for all other properties

    public function jsonSerialize(): array
    {
        // Use get_object_vars to get all public properties
        $vars = get_object_vars($this);
        // Filter out null values to keep the JSON clean
        return array_filter($vars, fn($value) => $value !== null);
    }
}
