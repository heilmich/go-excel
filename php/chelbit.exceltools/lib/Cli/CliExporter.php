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

    public function exportToFile(string $destPath, Schema $schema, ?string $templatePath = null): void
    {
        if (!$this->cliPath || !is_executable($this->cliPath)) {
            throw new RuntimeException("CLI path is not configured or not executable: {$this->cliPath}");
        }

        $payload = $schema->jsonSerialize();

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
