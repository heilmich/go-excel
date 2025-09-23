<?php
namespace Chelbit\Exceltools;

use Chelbit\Exceltools\Cli\CliClient;
use Chelbit\Exceltools\Cli\CliExporter;
use Chelbit\Exceltools\Data\Schema;
use Chelbit\Exceltools\Http\HttpClient;
use Chelbit\Exceltools\Enums\Mode;

/**
 * Единая точка входа для импорта/экспорта.
 * В зависимости от настроек модуля (service|cli) вызывает нужные реализации.
 */
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
            $this->importer = new CliClient($cliPath, Config::getCliTimeout());
            $this->exporter = new CliExporter($cliPath);
        } else {
            $httpClient = new HttpClient(Config::getServiceUrl(), Config::getAuthToken());
            $this->importer = $httpClient;
            $this->exporter = $httpClient;
        }
    }

    /** Импорт: генератор NDJSON. @param Schema|array|string $schema */
    public function parse(string $filePath, $schema): \Generator
    {
        if (!$this->importer) {
            throw new \RuntimeException('Importer is not configured');
        }
        return $this->importer->parse($filePath, Schema::fromMixed($schema));
    }

    /** Импорт: стримовый коллбек. Только для HTTP (CLI вариант уже потоковый через stdout). */
    public function parseStream(string $filePath, $schema, callable $onRow): void
    {
        if ($this->importer instanceof HttpClient) {
            $this->importer->parseStream($filePath, Schema::fromMixed($schema), $onRow);
            return;
        }

        // Для CLI используйте parse() и читайте генератор
        foreach ($this->parse($filePath, $schema) as $row) {
            $onRow($row);
        }
    }

    /** Экспорт в файл. @param Schema|array|string $schema */
    public function exportToFile(string $destPath, array $rows, $schema, array $formulas = [], array $options = [], ?string $templatePath = null): void
    {
        if (!$this->exporter) {
            throw new \RuntimeException('Exporter is not configured');
        }

        if ($this->exporter instanceof HttpClient) {
            $bytes = $this->exporter->export($rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
            $dir = dirname($destPath); if (!is_dir($dir)) { @mkdir($dir, 0775, true); }
            file_put_contents($destPath, $bytes);
        } else { // Must be CliExporter
            $this->exporter->exportToFile($destPath, $rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
        }
    }

    /** Экспорт в память (байты XLSX). */
    public function exportToBytes(array $rows, $schema, array $formulas = [], array $options = [], ?string $templatePath = null): string
    {
        if (!$this->exporter) {
            throw new \RuntimeException('Exporter is not configured');
        }

        if ($this->exporter instanceof HttpClient) {
            return $this->exporter->export($rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
        }

        // For CLI, we must write to a temp file and read it back
        $tmpFile = tempnam(sys_get_temp_dir(), 'exp_') . '.xlsx';
        $this->exporter->exportToFile($tmpFile, $rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
        $bytes = file_get_contents($tmpFile);
        @unlink($tmpFile);

        return (string)$bytes;
    }

    public function getMode(): string
    {
        return $this->mode;
    }
}
