<?php
$port = getenv('PORT') ?: '8000';
$host = $_SERVER['HTTP_HOST'] ?? 'unknown';
$path = $_SERVER['REQUEST_URI'] ?? '/';
?>
<!DOCTYPE html>
<html>
<head><title>PHP Server</title></head>
<body>
    <h1>PHP Built-in Server</h1>
    <p>Port: <?= htmlspecialchars($port) ?></p>
    <p>Host: <?= htmlspecialchars($host) ?></p>
    <p>Path: <?= htmlspecialchars($path) ?></p>
</body>
</html>
