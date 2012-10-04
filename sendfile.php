<?php

  include "init.inc";

  #import_request_variables("cpg","rvar_");
  extract($_REQUEST, EXTR_PREFIX_ALL|EXTR_REFS, 'rvar');

  header('Cache-Control: max-age=0, must-revalidate, no-cache, no-store');
  header('Content-Type: text/plain');

  $user = sanitize_word($rvar_user);
  $file = sanitize_path($rvar_file);

  if(!is_dir("$root_dir$user")) {
    header('HTTP/1.0 404 Not Found');
    die;
  }

  if(!is_file("$root_dir$user/$file")) {
    header('HTTP/1.0 404 Not Found');
    die;
  }

  readfile("$root_dir$user/$file");

?>
