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

    public function parse(string $filePath, Schema $schema): array
    {
        $ch = curl_init($this->baseUrl . '/api/parse');
        $fields = [
            'file' => new \CURLFile($filePath, 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', basename($filePath)),
            'schema' => json_encode($schema, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES),
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

    public function parseStream(string $filePath, Schema $schema, callable $onRow): void
    {
        // NOTE: The Go service no longer streams NDJSON for import.
        // This method's behavior is now inconsistent with the service.
        // A "proper" fix would be to remove this method, but for now we will
        // make it work by calling the regular parse and iterating over the result.
        $resultArray = $this->parse($filePath, $schema);
        foreach ($resultArray as $row) { // This will likely be just the top-level 'data' and 'errors' keys
            $onRow($row);
        }
    }

    public function export(Schema $schema, ?string $templatePath = null): string
    {
        $payload = $schema->jsonSerialize();
        $ch = curl_init($this->baseUrl . '/api/export');

        if ($templatePath && is_file($templatePath)) {
            $headers = $this->buildAuthHeaders();
            $headers[] = 'Content-Type: multipart/form-data';
            $fields = [
                'request' => json_encode($payload, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES),
                'template' => new \CURLFile($templatePath),
            ];
            curl_setopt($ch, CURLOPT_POSTFIELDS, $fields);
        } else {
            $headers = array_merge(['Content-Type: application/json'], $this->buildAuthHeaders());
            curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload, JSON_UNESCAPED_UNICODE|JSON_UNESCAPED_SLASHES));
        }

        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

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
