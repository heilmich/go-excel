<?php
namespace Chelbit\Exceltools\Data;

use Chelbit\Exceltools\Enums\Types;

final class Validator
{
    private array $custom = [];

    public function addCustom(string $field, callable $fn): void { $this->custom[$field] = $fn; }

    public function validateRow(array $row, Schema $schema): array
    {
        $errors = [];
        foreach ($schema->columns as $col) {
            $name = $col->name;
            $raw = $row[$name] ?? null;

            if ($col->required && ($raw === null || $raw === '')) {
                $errors[] = ['field'=>$name, 'msg'=>'required']; continue;
            }
            if ($raw === null) { continue; }
            $s = (string)$raw;

            if ($col->minLen !== null && mb_strlen($s) < $col->minLen) {
                $errors[] = ['field'=>$name, 'msg'=>'minLen '.$col->minLen];
            }
            if ($col->maxLen !== null && mb_strlen($s) > $col->maxLen) {
                $errors[] = ['field'=>$name, 'msg'=>'maxLen '.$col->maxLen];
            }
            if ($col->regex && !preg_match('/'.$col->regex.'/', $s)) {
                $errors[] = ['field'=>$name, 'msg'=>'regex '.$col->regex];
            }
            if ($col->enum && !in_array($s, $col->enum, true)) {
                $errors[] = ['field'=>$name, 'msg'=>'enum '.implode(',', $col->enum)];
            }
            if (in_array($col->type, [Types::INT, Types::FLOAT], true)) {
                $num = (float)str_replace(',', '.', $s);
                if ($col->min !== null && $num < $col->min) $errors[] = ['field'=>$name, 'msg'=>'min '.$col->min];
                if ($col->max !== null && $num > $col->max) $errors[] = ['field'=>$name, 'msg'=>'max '.$col->max];
            }
            if (isset($this->custom[$name]) && !($this->custom[$name])($raw)) {
                $errors[] = ['field'=>$name, 'msg'=>'custom'];
            }
        }
        return $errors;
    }
}
