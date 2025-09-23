<?php
namespace Chelbit\Exceltools;

use Chelbit\Exceltools\Cli\CliClient;
use Chelbit\Exceltools\Cli\CliExporter;
use Chelbit\Exceltools\Data\Schema;
use Chelbit\Exceltools\Http\HttpClient;
use Chelbit\Exceltools\Enums\Mode;
use Generator;
use RuntimeException;

final class Client
{
    private HttpClient|CliClient|null $importer = null;
    private HttpClient|CliExporter|null $exporter = null;
    private string $mode;

    public function __construct(?string $mode = null)
    {
        $this->mode = $mode ?? Config::getMode();
        if (!Mode::isValid($this->mode)) {
            throw new \InvalidArgumentException("Invalid mode: {$this->mode}");
        }

        if ($this->mode === Mode::CLI) {
            $cliPath = Config::getCliPath();
            // The timeout value was lost in previous refactoring, let's add it back.
            $timeout = Config::getCliTimeout() ?: 60;
            $this->importer = new CliClient($cliPath, $timeout);
            $this->exporter = new CliExporter($cliPath);
        } else {
            $httpClient = new HttpClient(Config::getServiceUrl(), Config::getAuthToken());
            $this->importer = $httpClient;
            $this->exporter = $httpClient;
        }
    }

    /**
     * Imports an Excel file and yields rows as associative arrays.
     * The schema should be configured for import using `forImport()`.
     */
    public function import(string $filePath, Schema $schema): Generator
    {
        if (!$this->importer) {
            throw new RuntimeException('Importer is not configured');
        }
        return $this->importer->parse($filePath, $schema);
    }

    /**
     * Exports data to an Excel file and saves it to the specified path.
     * The schema should be configured with export data using methods like `withRows()`, `addBlock()`, etc.
     */
    public function exportToFile(string $outputPath, Schema $schema, ?string $templatePath = null): void
    {
        if (!$this->exporter) {
            throw new RuntimeException('Exporter is not configured');
        }

        if ($this->exporter instanceof HttpClient) {
            $bytes = $this->exporter->export($schema, $templatePath);
            $dir = dirname($outputPath); if (!is_dir($dir)) { @mkdir($dir, 0775, true); }
            if (file_put_contents($outputPath, $bytes) === false) {
                throw new RuntimeException("Failed to write exported file to {$outputPath}");
            }
        } else { // Must be CliExporter
            $this->exporter->exportToFile($outputPath, $schema, $templatePath);
        }
    }

    /**
     * Exports data to an Excel file and returns its content as a byte string.
     */
    public function exportToBytes(Schema $schema, ?string $templatePath = null): string
    {
        if (!$this->exporter) {
            throw new RuntimeException('Exporter is not configured');
        }

        if ($this->exporter instanceof HttpClient) {
            return $this->exporter->export($schema, $templatePath);
        }

        // For CLI, we must write to a temp file and read it back
        $tmpFile = tempnam(sys_get_temp_dir(), 'export_') . '.xlsx';
        $this->exporter->exportToFile($tmpFile, $schema, $templatePath);
        $bytes = file_get_contents($tmpFile);
        @unlink($tmpFile);

        return $bytes;
    }
}
