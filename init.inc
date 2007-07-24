<?php

$root_dir = "./db/";

function sanitize_word($buf) {

  if( preg_match('/([A-Za-z0-9\.\-]+)/', $buf, $matches)) {
    return $matches[1];
  } else {
    return;
  }

}

function sanitize_path($buf) {

  $buf = preg_replace('/\.\./','',$buf);

  if( preg_match('/([A-Za-z0-9\.\-\/\_]+)/', $buf, $matches)) {
    return $matches[1];
  } else {
    return;
  }

}

function file_details($path,$file,$base = '') {
  $display = $file;

  if( preg_match('/([^:]+):(\d\d\d)/', $file, $matches)) {
    $display = $matches[1];
    $perms = $matches[2];
  }

  print $base;

  if(is_dir($path . $file)) {
    
    if(!isset($perms)) { $perms = 700; }

    print $display . "/\t";
    print $perms . "\t";
    print "*\t";
    print "\n";
  } else {
    if(!isset($perms)) { $perms = 600; }

    print $display . "\t";
    print $perms . "\t";
    print "*\t";
    print filesize($path . $file) . "\t";
    print md5_file($path . $file) . "\t";
    print "\n";
  }

  return;
}

function walk_dir($dir, $base = '') {

  if( $dh = opendir($dir) ) {

    while($file = readdir($dh)) {
      if($file != '.' and $file != '..' and $file != 'CVS') {
        if($base != '' and $file != 'bin/') {
          file_details("$dir/", $file, $base);
	}
	if(is_dir("$dir/$file")) {
	  walk_dir("$dir/$file","$base$file/");
	}
      }
    }
    closedir($dh);
  }

  return;
}

function custom_client_script($user) {

  $client_script = "./client/netskel";

  $buf = file_get_contents($client_script);

  $buf = preg_replace('/%HOST%/',$_SERVER["SERVER_NAME"],$buf);
  $buf = preg_replace('/%USER%/',$user,$buf);

  if(preg_match("/(.*)\/$user\/bin\/netskel/",$_SERVER["REQUEST_URI"],$matches)) {
    $buf = preg_replace('/%BASE%/',$matches[1],$buf);
  }
  if(preg_match("/(.*)\/$user\/netskel/",$_SERVER["REQUEST_URI"],$matches)) {
    $buf = preg_replace('/%BASE%/',$matches[1],$buf);
  }
  if(preg_match("/(.*)\/$user\/.netskeldb/",$_SERVER["REQUEST_URI"],$matches)) {
    $buf = preg_replace('/%BASE%/',$matches[1],$buf);
  }

  return $buf;

}

?>
