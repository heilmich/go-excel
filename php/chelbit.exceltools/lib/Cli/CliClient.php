<?php
namespace Chelbit\Exceltools\Cli;

use Chelbit\Exceltools\Data\Schema;
use Generator;
use RuntimeException;

final class CliClient
{
    private string $cliPath;
    private int $timeout;

    public function __construct(string $cliPath, int $timeout = 60)
    {
        if (!is_executable($cliPath)) {
            // It's better to let this fail at runtime than to prevent object creation,
            // as the path might be resolved later or might not be used at all.
        }
        $this->cliPath = $cliPath;
        $this->timeout = $timeout;
    }

    public function parse(string $filePath, Schema $schema): array
    {
        if (!is_readable($filePath)) {
            throw new RuntimeException("Import file is not readable: {$filePath}");
        }
        if (!$this->cliPath || !is_executable($this->cliPath)) {
            throw new RuntimeException("CLI path is not configured or not executable: {$this->cliPath}");
        }

        $schemaFile = tempnam(sys_get_temp_dir(), 'schema_import_');
        file_put_contents($schemaFile, json_encode($schema));

        $cmd = [
            escapeshellarg($this->cliPath),
            'import',
            '--schema', escapeshellarg($schemaFile),
            '--file', escapeshellarg($filePath),
        ];

        $output = shell_exec(implode(' ', $cmd));
        unlink($schemaFile);

        if ($output === null) {
            // This could indicate an error, check stderr if possible or rely on exit codes
            // For simplicity, we assume an empty array is a valid result for no output.
            return [];
        }

        return json_decode($output, true) ?? [];
    }

    private function popen(string $cmd): array
    {
        $descriptorspec = [
           0 => ["pipe", "r"], // stdin
           1 => ["pipe", "w"], // stdout
           2 => ["pipe", "w"], // stderr
        ];

        $process = proc_open($cmd, $descriptorspec, $pipes);

        if (!is_resource($process)) {
            throw new RuntimeException("Failed to open process with command: {$cmd}");
        }

        return ['process' => $process, 'stdin' => $pipes[0], 'stdout' => $pipes[1], 'stderr' => $pipes[2]];
    }

    private function pclose(array $process): void
    {
        $stderr = stream_get_contents($process['stderr']);
        fclose($process['stdin']);
        fclose($process['stdout']);
        fclose($process['stderr']);
        $exitCode = proc_close($process['process']);

        if ($exitCode !== 0) {
            throw new RuntimeException("CLI process exited with code {$exitCode}: {$stderr}");
        }
    }
}
