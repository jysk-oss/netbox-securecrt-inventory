# NetBox SecureCRT Inventory

Have you always wanted the ability to synchronize your NetBox devices with your SecureCRT client? Then this is for you!

This tool will automatically run in the background and perform periodic synchronization or run on-demand depending on your needs. It's highly flexible in terms of how the device structure is organized.

## Installation

1. Download the latest release zip file to your computer and unzip it.
2. Create a config file named `.securecrt-inventory.yaml` in your home directory (see below for a full config example):
   - Change `netbox_url`, and `netbox_token` as needed.
   - Update `overrides` as needed.
   - NB: It's also possible to create the config file by running the program as it will be created with defaults if missing.
3. Run the program. It should now start running as a systray program.
4. Optional: Set it to start automatically on Windows/macOS startup.

*Note:* On macOS, you might need to run `xattr -cr securecrt-inventory` to be able to execute it, as the binary is not signed. Alternatively, consider building the code yourself as a workaround.

## Templates and Expressions

The config supports two special types: templates and expressions. In this section, we'll cover the differences and how to use them.

### Templates

Templates provide a simple way to describe what value should be placed in a field. A good example is how it's used for the default path.

A template is a string with one or more `{}` placeholders inside. For example: `NetBox/{tenant_name}/{site_name}`.
If the tenant is "Example" and the site is "Test", the template would evaluate to `"NetBox/Example/Test`.

Templates have access to the following variables:
```
session_type: Either device or virtual_machine
credential: The default session credential name
path_template: The default path template
device_name_template: The default device name template
firewall_template: The default firewall template
connection_protocol_template: The default connection protocol template
device_name: Device name from NetBox
device_role: Device role name from NetBox
device_type: Device type name from NetBox
device_ip: Device IP without subnet/prefix
region_name: Region name from NetBox
tenant_name: Tenant name from NetBox
site_name: Site name from NetBox
site_group: Site Group name from NetBox
site_address: Site address from NetBox
```

### Expressions

Expressions are powered by https://expr-lang.org/ and should always start with `{{` and end with `}}`. They are used extensively to define overrides and manipulate the session output.

Here are a few sample expressions to get you started:
```
# Returns true if the NetBox site group is adm
{{ site_group == 'adm' }}

# Returns the value of the tag "connection_protocol" if found, otherwise "SSH"
{{ FindTag(device.Tags, 'connection_protocol') ?? 'SSH' }}

# Returns true if the device name ends with example.com
{{ device_name endsWith '.example.com' }}
```

Expressions have access to the same variables as templates, but they can also access the following:
```
device: The device object (go struct, most fields are CamelCase, ex: device.Tags)
site: The site object  (go struct, most fields are CamelCase, ex: site.Slug)
```

Expressions have access to all expr functions and the following:
```
FindTag(<tags>, <tag_name)
```

### Debug
It's possible to debug expressions and templates, by enabling debug in the config file and examining the log file. The log file can be opened by clicking the icon and selecting "Open Log", when debug is enabled all variables will be output together with templates, and result.

**IMPORTANT**: This should not be enabled always as the log file is not rotated, and debug WILL output a lot of data.

## Config Example

```
# ERROR/DEBUG/INFO, default is ERROR. DEBUG logs a lot and should not be used in day-to-day operations as the log is not cleared.
log_level: ERROR 
netbox_url: <netbox_url>
netbox_token: <netbox_token>
root_path: NetBox

# Enable/Disable periodic sync (note: SecureCRT needs to be restarted for changes to take effect)
periodic_sync_enable: true
periodic_sync_interval: 120

# Session settings
session:
  # path: is the default session path template
  path: "{tenant_name}/{region_name}/{site_name}/{device_role}"
  # device_name: allows you to override the device name at a global level; supports templates and expressions
  device_name: "{device_name}"

  # Global Session Options
  session_options:
    # Allows you to override the connection protocol; supports templates and expressions
    connection_protocol: "{{ FindTag(device.Tags, 'connection_protocol') ?? 'SSH' }}"
    # Set default credentials; they should be defined in SecureCRT beforehand under "Preferences -> General -> Credentials"
    credential: <username>
    # Set a firewall; supports templates and expressions
    firewall: "{{ FindTag(device.Tags, 'connection_firewall') ?? ''None'' }}"

  # Overrides based on conditions
  # target can be one of: path, device_name, description, connection_protocol, credential, firewall
  # condition should always be an expression that evaluates to true or false
  # value is what to replace the target with; it can be a template or expression that returns a value
  overrides:
    - target: path
      condition: "{{ site_group == 'adm' }}"
      value: _Stores/{region_name}/{site_name}

    - target: path
      condition: "{{ device_type == 'virtual_machine' }}"
      value: _Servers/{region_name}
    
    # device_name override use cases include removing domain names, extra values like .1, and so on
    # note that this example could also be done with device_name and just using replace
    - target: device_name
      condition: "{{ device_name endsWith '.1' }}"
      value: "{{ replace(device_name, '.1', '') }}"
```

## Development
Pull requests and issues are welcome.

A VSCode launch file has been included for debugging the code directly.