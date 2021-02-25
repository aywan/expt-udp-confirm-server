<?php
const TESTING_PACKET_SIZE = 1024 * 64;

const MAX_PACKET_SIZE=1400;
const SERVER = 1;

$pid = getmypid();
if (false === $pid) {
    fprintf(STDERR, "false pid");
    exit(1);
}
echo $pid, ' / ', SERVER, PHP_EOL; 

$data = '';

for ($i=0; $i < TESTING_PACKET_SIZE; $i++) {
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
if(! $sock) {
	$errorcode = socket_last_error();
    $errormsg = socket_strerror($errorcode);
    
    die("Couldn't create socket: [$errorcode] $errormsg \n");
}

$parts = [];
for($i = 0; $i < $packets; $i++) {
    $number = $i << 16 | $packets;
    $packetData = substr($data, $i * PACKET_SIZE, PACKET_SIZE - 1);
    $packetDataLen = strlen($packetData);
    $crc = crc32($packetData);
    $packet = pack('VVvvVV', $pid, SERVER, $i, $packets, $crc, $packetDataLen) . $packetData;

    if( ! socket_sendto($sock, $packet , strlen($packet) , 0 , $server , $port)) {
		$errorcode = socket_last_error();
		$errormsg = socket_strerror($errorcode);
		
		die("Could not send data: [$errorcode] $errormsg \n");
	}

    echo $i, '/', $packets, ' ';

    if(socket_recv ($sock, $reply , MAX_PACKET_SIZE , MSG_WAITALL ) === FALSE) {
		$errorcode = socket_last_error();
		$errormsg = socket_strerror($errorcode);
		
		die("Could not receive data: [$errorcode] $errormsg \n");
	}

    $parts = unpack('v', $reply);
    $code = $parts[1];

    $map = [
        1 => 'OK',
        2 => 'WRONG_LENGTH',
        4 => 'WRONG_CRC',
    ];

    echo $map[$code] ?? 'unknown', PHP_EOL;
}

