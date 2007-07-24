<?php

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
        file_details("$dir/", $file, $base);
	if(is_dir("$dir/$file")) {
	  walk_dir("$dir/$file","$base$file/");
	}
      }
    }
    closedir($dh);
  }

  return;
}

?>
