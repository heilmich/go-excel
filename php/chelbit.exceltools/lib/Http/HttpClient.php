<?php
namespace Chelbit\Exceltools\Http;

use Chelbit\Exceltools\Data\Schema;

final class HttpClient
{
    private string $baseUrl;
    private ?string $token;

    public function __construct(string $baseUrl, ?string $token=null)
    {
        $this->baseUrl = rtrim($baseUrl, '/');
        $this->token = $token;
    }

    public function ping(): bool
    {
        $ch = curl_init($this->baseUrl . '/health');
        curl_setopt_array($ch, [CURLOPT_RETURNTRANSFER=>true, CURLOPT_TIMEOUT=>10]);
        $data = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        return ($code === 200) && is_string($data);
    }

    /** @param Schema|array|string $schema */
    public function parse(string $filePath, $schema): array
    {
        $schemaObj = Schema::fromMixed($schema);

        $ch = curl_init($this->baseUrl . '/api/parse');
        $fields = [
            'file' => new \CURLFile($filePath, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', basename($filePath)),
            'schema' => json_encode($schemaObj, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES),
        ];
        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_HTTPHEADER => $this->buildAuthHeaders(),
            CURLOPT_POSTFIELDS => $fields,
            CURLOPT_RETURNTRANSFER => true,
        ]);

        $response = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($code !== 200) {
            throw new \RuntimeException('Импорт: ошибка сервиса (' . $code . '): ' . $response);
        }

        return json_decode($response, true) ?? [];
    }

    /** @param Schema|array|string $schema */
    public function parseStream(string $filePath, $schema, callable $onRow): void
    {
        $schemaObj = Schema::fromMixed($schema);

        $ch = curl_init($this->baseUrl . '/api/parse');
        $fields = [
            'file' => new \CURLFile($filePath, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', basename($filePath)),
            'schema' => json_encode($schemaObj, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES),
        ];
        curl_setopt_array($ch, [
            CURLOPT_POST => true,
            CURLOPT_HTTPHEADER => $this->buildAuthHeaders(),
            CURLOPT_POSTFIELDS => $fields,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_WRITEFUNCTION => function($ch, $data) use ($onRow) {
                static $buf = '';
                $buf .= $data;
                while (($pos = strpos($buf, "\n")) !== false) {
                    $line = trim(substr($buf, 0, $pos));
                    $buf = substr($buf, $pos+1);
                    if ($line === '') continue;
                    $row = json_decode($line, true);
                    if (is_array($row)) { $onRow($row); }
                }
                return strlen($data);
            }
        ]);
        $res = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        if ($code !== 200 && $res === false) {
            throw new \RuntimeException('Импорт (stream): ошибка сервиса (' . $code . ')');
        }
    }

    public function export(Schema $schema, ?string $templatePath = null): string
    {
        $payload = $schema->jsonSerialize();

        if ($templatePath && is_file($templatePath)) {
            $ch = curl_init($this->baseUrl . '/api/export');
            $headers = $this->buildAuthHeaders();
            $headers[] = 'Content-Type: multipart/form-data';
            $fields = [
                'request' => json_encode($payload),
                'template' => new \CURLFile($templatePath),
            ];
            curl_setopt_array($ch, [
                CURLOPT_POST => true,
                CURLOPT_HTTPHEADER => $headers,
                CURLOPT_POSTFIELDS => $fields,
                CURLOPT_RETURNTRANSFER => true,
            ]);
        } else {
            $ch = curl_init($this->baseUrl . '/api/export');
            $headers = array_merge(['Content-Type: application/json'], $this->buildAuthHeaders());
            curl_setopt_array($ch, [
                CURLOPT_POST => true,
                CURLOPT_HTTPHEADER => $headers,
                CURLOPT_POSTFIELDS => json_encode($payload, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES),
                CURLOPT_RETURNTRANSFER => true,
            ]);
        }

        $data = curl_exec($ch);
        $code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        if ($code !== 200) {
            throw new \RuntimeException('Экспорт: ошибка сервиса (' . $code . '): ' . $data);
        }
        return (string)$data;
    }

    private function buildAuthHeaders(): array
    {
        $h = [];
        if ($this->token) { $h[] = 'Authorization: Bearer ' . $this->token; }
        return $h;
    }
}
