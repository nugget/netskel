<?php

  include "init.inc";

  import_request_variables("cpg","rvar_");

  header('Cache-Control: max-age=0, must-revalidate, no-cache, no-store');
  header('Content-Type: text/plain');

  $user = sanitize_word($rvar_user);
  $root_dir = "./db/";

  if(!is_dir("$root_dir$user")) {
    header('HTTP/1.0 404 Not Found');
    die;
  }

  print "#\n";
  print "# .netskeldb for $user\n";
  print "#\n";
  print "# Generated " . date("d-M-Y @ H:m T") . " by " . $_SERVER['HTTP_HOST'] . "\n";
  print "#\n";

  walk_dir("$root_dir$user", '');

?>
