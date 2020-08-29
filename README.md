## About

The check plugin **check\_masifupgrader\_agent** monitors
the *Masif Upgrader agent*, a component of *Masif Upgrader*.

Consult Masif Upgrader's [manual] on its purpose
and the agent's role in its architecture.

## Usage

The binaries of the [Debian packages]
take the following CLI arguments and no environment variables:

```
$ /usr/lib/nagios/plugins/check_masifupgrader_agent \
-restsock /var/run/masif-upgrader-agent/rest.s \
[-resptime-(warn|crit) THRESHOLD] \
[-(query|install|update|configure|remove|purge|error)-(1m|5m|15m)-(warn|crit) THRESHOLD]
```

THRESHOLD specifies an alert threshold range
conforming to the [Nagio$ check plugin API],
e.g. `-error-5m-warn @~:42` warns if there occurred <= 42 errors
during the last 5 minutes.

### Legal info

To print the legal info, execute the plugin in a terminal:

```
$ /usr/lib/nagios/plugins/check_masifupgrader_agent
```

In this case the program will always terminate with exit status 3 ("unknown")
without actually checking anything.

### Testing

If you want to actually execute a check inside a terminal,
you have to connect the standard output of the plugin to anything
other than a terminal – e.g. the standard input of another process:

```
$ /usr/lib/nagios/plugins/check_masifupgrader_agent |cat
```

In this case the exit code is likely to be the cat's one.
This can be worked around like this:

```
bash $ set -o pipefail
bash $ /usr/lib/nagios/plugins/check_masifupgrader_agent |cat
```

### Actual monitoring

Just integrate the plugin into the monitoring tool of your choice
like any other check plugin. (Consult that tool's manual on how to do that.)
It should work with any monitoring tool
supporting the [Nagio$ check plugin API].

Limitations:

* check\_masifupgrader\_agent must be run on the host to be checked –
  either with an agent of your monitoring tool or by SSH.
  Otherwise it will check the host your monitoring tool runs on.
* **The user check\_masifupgrader\_agent runs as must be a member
  of the group masif-upgrader-agent**, i.e.:
  `usermod -aG masif-upgrader-agent nagios`
  Don't forget to restart your monitoring tool's service if any.

#### Icinga 2

The [Debian packages] ship the [check command definition] for [Icinga 2].
This repository ships a [service template] and a [host example] as well.

The service definition will work in both correctly set up [Icinga 2 clusters]
and Icinga 2 instances not being part of any cluster
as long as the [hosts] are named after the [endpoints].

[manual]: https://github.com/masif-upgrader/manual
[Debian packages]: https://github.com/masif-upgrader/check_masifupgrader_agent/releases
[Nagio$ check plugin API]: https://nagios-plugins.org/doc/guidelines.html#AEN78
[check command definition]: ./icinga2/check_masifupgrader_agent.conf
[Icinga 2]: https://www.icinga.com/docs/icinga2/latest/doc/01-about/
[service template]: ./icinga2/check_masifupgrader_agent-service.conf
[host example]: ./icinga2/check_masifupgrader_agent-host.conf
[Icinga 2 clusters]: https://www.icinga.com/docs/icinga2/latest/doc/06-distributed-monitoring/
[hosts]: https://www.icinga.com/docs/icinga2/latest/doc/09-object-types/#host
[endpoints]: https://www.icinga.com/docs/icinga2/latest/doc/09-object-types/#endpoint
