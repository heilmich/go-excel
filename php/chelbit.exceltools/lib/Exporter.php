<?php
namespace Chelbit\Exceltools;

final class Exporter
{
    /** @param Schema|array|string $schema */
    public static function exportBytes(array $rows, $schema, array $formulas = [], array $options = [], ?string $templatePath = null): string
    {
        $client = new Client();
        return $client->exportToBytes($rows, $schema, $formulas, $options, $templatePath);
    }

    /** @param Schema|array|string $schema */
    public static function exportToFile(string $destPath, array $rows, $schema, array $formulas = [], array $options = [], ?string $templatePath = null): void
    {
        $client = new Client();
        $client->exportToFile($destPath, $rows, $schema, $formulas, $options, $templatePath);
    }
}
