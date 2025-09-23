<?php
namespace Chelbit\Exceltools\Cli;

use Chelbit\Exceltools\Data\Schema;
use RuntimeException;

final class CliExporter
{
    private string $cliPath;

    public function __construct(string $cliPath)
    {
        $this->cliPath = $cliPath;
    }

    public function exportToFile(string $destPath, array $rows, Schema $schema, array $formulas = [], array $options = [], ?string $templatePath = null): void
    {
        if (!$this->cliPath || !is_executable($this->cliPath)) {
            throw new RuntimeException("CLI path is not configured or not executable: {$this->cliPath}");
        }

        // Construct the full request payload, just like the HTTP client does.
        $payload = [
            'sheet' => $schema->sheet ?? 'Sheet1',
            'headerRow' => $schema->headerRow,
            'startRow' => $schema->startRow,
            'columns' => array_map(fn($c) => $c->jsonSerialize(), $schema->columns),
            'rows' => $rows,
            'formulas' => $formulas,
            'options' => $options,
        ];

        $requestFile = tempnam(sys_get_temp_dir(), 'request_export_');
        file_put_contents($requestFile, json_encode($payload));

        // The Go CLI needs to be updated to accept a single --request file.
        $cmd = [
            escapeshellarg($this->cliPath),
            'export',
            '--request', escapeshellarg($requestFile),
            '--output', escapeshellarg($destPath),
        ];

        if ($templatePath) {
            $cmd[] = '--template';
            $cmd[] = escapeshellarg($templatePath);
        }

        $this->executeCommand(implode(' ', $cmd));

        unlink($requestFile);
    }

    private function executeCommand(string $cmd): void
    {
        exec($cmd . ' 2>&1', $output, $return_var);
        if ($return_var !== 0) {
            throw new RuntimeException("CLI export execution failed: " . implode("\n", $output));
        }
    }
}
