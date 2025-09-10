# Brother Cert Changelog

## [v0.3.0] - 2025-09-09

Many items were changed or overhauled in this version. Possibly the
most important change comes from @wesleykirkland. The app now scans
the printer's login page to identify the correct field name for the
login password field. This should greatly expand compatibility with
other printer models. Other field names appear not to change with
different models, so hopefully this one change is enough to bring
wider support for many.

- Add universal login scan to detect the correct name of the password
  field for printer login.
- Add workaround when a certificate has no Common Name. This is
  considered a more "modern" standard (e.g., the Let's Encrypt
  `tlsserver` profile) and is now supported.
- Add check to see if the specified certificate is already in use
  by the printer. If so, abort the install.
- Modify various regex patterns to make them more robust.
- Fix pem parsing function. The logic was not correct for files
  that didn't contain a 2nd certificate as part of the chain.
- Ensure the correct CSRF Token is always selected by the relevant
  regex.
- Change cert id into a string (instead of int). This eliminates
  unneeded type conversions and also is more flexible if some models
  use strings for their IDs instead of just numbers.
- Improve a number of error messages.
- Move User-Agent specification into a custom transport.
- Simplify build system and make it OS agnostic.
- Update all dependencies, including Go 1.25.1.
- Update README with more details about the application, usage, and
  how to build.


## [v0.2.1] - 2024-03-06

Update to Go 1.22.1, which includes some security fixes.


## [v0.2.0] - 2024-02-07

Initial release.
