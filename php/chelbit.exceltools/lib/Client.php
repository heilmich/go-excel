<?php
namespace Chelbit\Exceltools;

/**
 * Единая точка входа для импорта/экспорта.
 * В зависимости от настроек модуля (service|cli) вызывает нужные реализации.
 */
final class Client
{
    private ?HttpClient $http = null;
    private ?CliClient $cli = null;
    private ?CliExporter $cliExporter = null;
    private string $mode;

    public function __construct()
    {
        $this->mode = Config::getMode();
        if ($this->mode === 'cli') {
            $cliPath = Config::getCliPath();
            // Assuming one binary for now. The user docs were ambiguous.
            $this->cli = new CliClient($cliPath, Config::getCliTimeout());
            $this->cliExporter = new CliExporter($cliPath);
        } else {
            $this->http = new HttpClient(Config::getServiceUrl(), Config::getAuthToken());
        }
    }

    /** Импорт: генератор NDJSON. @param Schema|array|string $schema */
    public function parse(string $filePath, $schema): \Generator
    {
        if ($this->http) {
            return $this->http->parse($filePath, Schema::fromMixed($schema));
        }
        if ($this->cli) {
            return $this->cli->parse($filePath, Schema::fromMixed($schema));
        }
        throw new \RuntimeException('Некорректная конфигурация клиента');
    }

    /** Импорт: стримовый коллбек. Только для HTTP (CLI вариант уже потоковый через stdout). */
    public function parseStream(string $filePath, $schema, callable $onRow): void
    {
        if ($this->http) {
            $this->http->parseStream($filePath, Schema::fromMixed($schema), $onRow);
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
        if ($this->http) {
            $bytes = $this->http->export($rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
            $dir = dirname($destPath); if (!is_dir($dir)) { @mkdir($dir, 0775, true); }
            file_put_contents($destPath, $bytes);
            return;
        }
        if ($this->cliExporter) {
            $this->cliExporter->exportToFile($destPath, $rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
            return;
        }
        throw new \RuntimeException('Некорректная конфигурация клиента');
    }

    /** Экспорт в память (байты XLSX). */
    public function exportBytes(array $rows, $schema, array $formulas = [], array $options = [], ?string $templatePath = null): string
    {
        if ($this->http) {
            return $this->http->export($rows, Schema::fromMixed($schema), $formulas, $options, $templatePath);
        }

        $tmp = tempnam(sys_get_temp_dir(), 'exp_') . '.xlsx';
        $this->exportToFile($tmp, $rows, $schema, $formulas, $options, $templatePath);
        $data = file_get_contents($tmp); @unlink($tmp);
        return (string)$data;
    }

    public function getMode(): string
    {
        return $this->mode;
    }
}
