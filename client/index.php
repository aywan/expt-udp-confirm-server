<?php

const TESTING_PACKET_SIZE = 1024 * 64;

const MAX_PACKET_SIZE = 512;
$serverId = 1 << 24 | random_int(0, 1 << 24);

$pid = getmypid();
if (false === $pid) {
    fprintf(STDERR, "false pid");
    exit(1);
}
echo $pid, ' / ', $serverId, PHP_EOL;

$data = '';

for ($i = 0; $i < TESTING_PACKET_SIZE; $i++) {
    $data .= random_bytes(1);
}
$dataCrc = pack('V', crc32($data));
echo 'Data crc: ', bin2hex($dataCrc), PHP_EOL;
$data .= $dataCrc;

/**
 * packet structure
 * [0-31] - pid
 * [32-63] - server
 * [64-79] - packet number
 * [80-95] - packets count
 * [96-127] - crc32
 * [128-159] - length
 * [160-MPS] - data
 */

const HEADER_SIZE = 160 / 8;
const PACKET_SIZE = MAX_PACKET_SIZE - HEADER_SIZE;

$dataLen = strlen($data);
$packets = ceil($dataLen / PACKET_SIZE);

echo 'data size: ', $dataLen, PHP_EOL;
echo 'total pakets: ', $packets, PHP_EOL;

$server = '127.0.0.1';
$port = '18086';
$sock = socket_create(AF_INET, SOCK_DGRAM, 0);
if (! $sock) {
    $errorCode = socket_last_error();
    $errorMsg = socket_strerror($errorCode);

    die("Couldn't create socket: [$errorCode] $errorMsg \n");
}

$parts = [];
for ($i = 0; $i < $packets; $i++) {
    $packetData = substr($data, $i * PACKET_SIZE, PACKET_SIZE - 1);
    $packetDataLen = strlen($packetData);
    $crc = crc32($packetData);
    $packet = pack('VVvvVV', $pid, $serverId, $i, $packets, $crc, $packetDataLen) . $packetData;

    $parts[$i] = [
        'packet' => $packet,
        'len' => strlen($packet),
        'crc' => $crc % $pid,
    ];
}
$i = 0;
while (! empty($parts)) {

    if (isset($parts[$i])) {
        $res = socket_sendto($sock, $parts[$i]['packet'], $parts[$i]['len'], 0, $server, $port);
        if (! $res) {
            $errorCode = socket_last_error();
            $errorMsg = socket_strerror($errorCode);

            die("Could not send data: [$errorCode] $errorMsg \n");
        }
    }

    do {
        $bytes = socket_recv($sock, $reply, MAX_PACKET_SIZE, MSG_DONTWAIT);
        if ($bytes === 8) {
            $v = unpack('vstate/vid/Vcrc', $reply);
            if(1 !== $v['state']) {
                continue;
            }
            if (! isset($parts[$v['id']]) || $parts[$v['id']]['crc'] !== $v['crc']) {
                continue;
            }
            unset($parts[$v['id']]);
        }
    } while ($bytes > 0);


    $i++;
    if ($i >= $packets) {
        $i %= $packets;
    }
}
