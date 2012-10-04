<?php

  include "init.inc";

  extract($_REQUEST, EXTR_PREFIX_ALL|EXTR_REFS, 'rvar');


  header('Cache-Control: max-age=0, must-revalidate, no-cache, no-store');
  header('Content-Type: text/plain');

  $user = sanitize_word($rvar_user);
  $client_script = "./client/netskel";

  if(!is_dir("$root_dir$user")) {
    header('HTTP/1.0 404 Not Found');
    die;
  }

  $buf = custom_client_script($user);

  header('Content-type: text/plain');
#  header('Content-Disposition: attachment; filename=netskel');
#  header('Content-Description: netskel client script');
  header('Cache-Control: no-store, no-cache, must-revalidate');
  header('Cache-Control: post-check=0, pre-check=0', false);
  header('Pragma: no-cache');

  print $buf;

?>
