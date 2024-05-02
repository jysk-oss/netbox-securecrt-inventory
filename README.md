# NetBox SecureCRT Inventory
Have you always wanted the ability to synchronize your NetBox devices to your SecureCRT client? Then this is for you!

It will automatically run in the background and do a perodic synchronization, or run on-demand depending on the need. It's highly fleksible in regards to how the devices structure is made.

## Installation
1. Download the latest release zip file to your computer, and unzip it.
2. Create a config file named .securecrt-inventory.yaml in your home directory (see below for a full config example)
   - Change default_credential, netbox_url, and netbox_token as needed
   - Update session_path overwrites / name overwrites as needed
3. Run the program. It should now start running as a systray program
4. Optional: Set it to start automatically on windows/mac startup

*Note:* On OSX you might need to do `xattr -cr securecrt-inventory` to be able to run it, this is because the binary is not signed. Alternatively, consider building the code yourself as a workaround.

## Config
```
netbox_url: <netbox_url>
netbox_token: <netbox_token>
root_path: NetBox

# Name overwrites can be used to change the name of the session
# Typical usecases is to remove domain names, extra values like .1 and so on
name_overwrites:
  - regex: "\\.1$"
    value: ""

# Session Path defines where a session is saved
# There's a default template, but it's also possible to create overwrites based on the following keys:
# site_group, type, tenant_name, region_name, site_name, device_name
session_path:
    template: '{tenant_name}/{region_name}/{site_name}/{device_role}'
    overwrites:
      - key: site_group
        value: Test Sites
        template: _TEST/{region_name}/{site_name}
      - key: type
        value: virtual_machine
        template: _Servers/{region_name}
# Enable / Disable periodic sync (note: SecureCRT needs to be restart for changes to take affect)
periodic_sync_enable: true
periodic_sync_interval: 120

# Default credentails to use, they should be defined in SecureCRT beforehand under "Preferences -> General -> Credentials"
default_credential: <username>
```

## Development
PR's, and issues are welcome.

A VSCode launch file has been included for debugging the code directly.
