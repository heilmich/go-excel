<?php
namespace Chelbit\Exceltools\Data;

use JsonSerializable;

/**
 * Represents a Table, which is a Block that shows headers by default.
 */
final class Table extends Block implements JsonSerializable
{
    public function __construct(string $startCell, array $columns, array $data)
    {
        parent::__construct($startCell, $columns, $data);
        $this->showHeaders = true; // Default for a Table is true
    }
}
