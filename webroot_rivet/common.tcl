namespace eval ::netskel {

	proc require_values {arrayVar args} {
		upvar 1 $arrayVar response

		foreach k $args {
			if {![info exists response($k)]} {
				::netskel::deny_access
			}
		}
	}

	proc deny_access {} {
		headers numeric 403
		abort_page
	}

	proc dbdir {user} {
		return "/usr/local/htdocs/netskel/db/$user"
	}

	proc custom_client_script {user} {
		set client_script "../client/netskel"
		set buf [read_file $client_script]

		set buf [regsub -all {%HOST%}  $buf [env SERVER_NAME]]
		set buf [regsub -all {%USER%}  $buf $user]
		set buf [regsub -all {%PROTO%} $buf https]

		if {[regexp "/(.*)\/$user\/bin\/netskel" [env REQUEST_URI] _ match]} {
			set buf [regsub -all {%BASE%} $buf "/$match"]
		} elseif {[regexp "/(.*)\/$user\/.netskeldb" [env REQUEST_URI] _ match]} {
			set buf [regsub -all {%BASE%} $buf "/$match"]
		} else {
			set buf [regsub -all {%BASE%} $buf "HEY!  [env REQUEST_URI]"]
		}

		return $buf
	}

	proc file_details {path file {base ""}} {
		set retbuf ""

		set display $file

		if {[regexp {([^:]+):(\d\d\d)} $file _ l_display l_perms]} {
			set display $l_display
			set perm $l_perms
		}

		if {[file isdirectory [file join $path $file]]} {
			if {![info exists perms]} {
				set perms 700
			}
			set retbuf "[file join ${base} ${display}]/\t$perms\t*\t"
		} else {
			set perms [file attributes $file -permissions]
			set size [file size $file]
			set hash [string tolower [::md5::md5 -hex -file $file]]
			set retbuf  "[file join ${base} ${display}]\t$perms\t*\t$size\t$hash"
		}

		return $retbuf
	}

	proc walk_dir {dbdir {base ""}} {
		cd $dbdir
		foreach file [concat [glob -nocomplain ".*"] [glob -nocomplain "*"]] {
			cd $dbdir
			if {$file ne "." && $file ne ".."} {
				puts [::netskel::file_details $dbdir $file $base]
				if {[file isdirectory $file]} {
					::netskel::walk_dir [file join $dbdir $file] [file join $base $file]
				}
			}
		}
	}

}
