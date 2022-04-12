<?php

$configs = include('config.php');

$conn = new mysqli($configs['dbhost'], $configs['dbuser'], $configs['dbpw']);
if ($conn->connect_error) {
    die($conn->connect_error);
}
$conn->select_db($configs['dbname']);

if ($_GET["action"] == "get_tags") {
    $result = $conn->query("SELECT name, id FROM tags");
    if ($result == false) {
        die($conn->error);
    }

    echo "[";
    if ($result->num_rows > 0) {
        $num = $result->num_rows;
        while ($row = $result->fetch_assoc()) {
           
            echo "{\"Name\":\"".$row["name"]."\",\"Id\":\"".$row["id"]."\"}";
            if (--$num > 0) {
                echo ",";
            }
        }
    }
    echo "]";
}
else if ($_GET["action"] == "find_tag_by_name") {
    $safe_name =  $conn->real_escape_string($_GET['name']);

    $result = $conn->query("SELECT id FROM tags WHERE name='$safe_name'");
    if ($result == false) {
        die($conn->error);
    }

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        echo $row["id"];
    }
}
else if ($_GET["action"] == "update_tag") {
    $safe_name =  $conn->real_escape_string($_GET['name']);
    $safe_new_id =  $conn->real_escape_string($_GET['new_id']);

    $result = $conn->query("INSERT INTO tags (name, id) VALUES ('$safe_name','$safe_new_id') ON DUPLICATE KEY UPDATE id='$safe_new_id'");
    if ($result == false) {
        die($conn->error);
    }
}
else if ($_GET["action"] == "find_entry") {
    $safe_id =  $conn->real_escape_string($_GET['id']);

    $result = $conn->query("SELECT base_id FROM entries WHERE id='$safe_id' LIMIT 1");
    if ($result == false) {
        die($conn->error);
    }

    if ($result->num_rows > 0) {
        $row = $result->fetch_assoc();
        echo $row["base_id"];
    }
}
else if ($_GET["action"] == "add_entry") {
    $safe_id =  $conn->real_escape_string($_GET['id']);
    $safe_base_id =  $conn->real_escape_string($_GET['base_id']);

    $result = $conn->query("INSERT INTO entries (id, base_id) VALUES ('$safe_id','$safe_base_id')");
    if ($result == false) {
        die($conn->error);
    }
}

?>