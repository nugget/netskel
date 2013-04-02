<?php

  include "init.inc";

  extract($_REQUEST, EXTR_PREFIX_ALL|EXTR_REFS, 'rvar');

  header('Cache-Control: max-age=0, must-revalidate, no-cache, no-store');
  header('Content-Type: text/plain');

  $user = sanitize_word($rvar_user);

  if(!is_dir("$root_dir$user")) {
    header('HTTP/1.0 404 Not Found');
    die;
  }

  date_default_timezone_set('UTC');
  print "#\n";
  print "# .netskeldb for $user\n";
  print "#\n";
  print "# Generated " . date("d-M-Y @ H:i T") . " by " . $_SERVER['HTTP_HOST'] . "\n";
  print "#\n";

  $buf = custom_client_script($user);
  $scriptsize = strlen($buf);
  $scriptmd5 = md5($buf);

  print "bin/\t700\t*\n";
  print "bin/netskel\t700\t*\t$scriptsize\t$scriptmd5\t\n";

  walk_dir("$root_dir$user", '');
?>
