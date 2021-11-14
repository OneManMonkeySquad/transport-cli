# CLI for binary releases and patching 🚢

CLI tool to distribute (incremental) releases to users. Different release streams are supported (via *tags*).

To host the files you need a webserver. SFTP is used to upload the files while HTTP is used to download them.


## How to use
First you need to copy *transport.toml.example* to *transport.toml* and insert your data. Note that you can use the *local* backend to get a feel for how transport works.

Once configured, create a base patch:
```
tp base C:/path_to_app
base:271df855-2056-4bd5-b6ba-f7d14857820e
```
Upload the patch:
```
tp commit latest base:271df855-2056-4bd5-b6ba-f7d14857820e
```

Test the download:
```
tp restore latest C:/app_release_test
```

Create patch:
```
tp patch latest C:/path_to_app
patch:e310512b-fefe-42d6-90b5-bca96730411a
```

Upload the patch:
```powershell
tp commit latest patch:e310512b-fefe-42d6-90b5-bca96730411a
```


## Reference
```powershell
tp base {dir}
```
Create a base patch with all files included in the patch. The command will return the *patch ID* of the newly created patch. It does **not** actually create a release or upload anything.

```powershell
tp patch {tag} {dir}
```
Create an incremental patch with file differences included in the patch. The command will return the *patch ID* of the newly created patch. It does **not** actually create a release or upload anything.

```powershell
tp commit {tag} {patch_guid}
```
Upload the patch and make this the newest release for the given tag.

```powershell
tp restore {tag} {dir}
```
Applies changes to the directory until it matches the file states stated by the tag. This installs or updates the target software.

```powershell
tp tags
```
Print existing tags. F.i. stable, development, latest, ...


## Development status
Basic workflow is working. Files are only changed when needed (MD5 hash). File deletions are included too. File contents are not patched incrementally yet and are included directly in the patch json.

Do we need a self updater?
Do we need an example gui?