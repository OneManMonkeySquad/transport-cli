# CLI for binary releases and patching 🚢

CLI tool to distribute (incremental) releases to users. Different release streams are supported (via *tags*).

To host the files you need a webserver. SFTP is used to upload the files while HTTP is used to download them.


## How to use
First you need to copy *transport.toml.example* to *transport.toml* and insert your data. Note that you can use the *local* backend to get a feel for how transport works.

Once configured, create a base patch:
```
./transport-cli base C:/path_to_app
```
Upload the patch:
```
./transport-cli commit latest
```

Test the download:
```
./transport-cli restore latest C:/app_release_test
```

Create patch:
```
./transport-cli patch latest C:/path_to_app
```

Upload the patch:
```powershell
./transport-cli commit latest
```


## Reference
```powershell
./transport-cli base {dir}
```
Create a base patch with all files included in the patch. The command will return the *patch ID* of the newly created patch. It does **not** actually create a release or upload anything.

```powershell
./transport-cli patch {tag} {dir}
```
Create an incremental patch with file differences included in the patch. The command will return the *patch ID* of the newly created patch. It does **not** actually create a release or upload anything.

```powershell
./transport-cli commit {tag} {patch_guid}
```
Upload the patch and make this the newest release for the given tag.

```powershell
./transport-cli restore {tag} {dir}
```
Applies changes to the directory until it matches the file states stated by the tag. This installs or updates the target software.

```powershell
./transport-cli tags
```
Print existing tags. F.i. stable, development, latest, ...


## Development status
Basic workflow is working. Files are only changed when needed (SHA256 hash). File deletions are included too. File contents are not patched incrementally yet. File blobs are zlib compressed.

Do we need a self updater?
Do we need an example gui?