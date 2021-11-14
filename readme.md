# CLI for binary releases and patching

CLI tool to distribute (incremental) releases to users. Different release streams are supported (via *tags*).

To host the files you need a webserver. SFTP is used to upload the files while HTTP is used to download them.


## How to use
```powershell
tp base {dir}
tp base path_to_app/
```
Create a base patch with all files included in the patch. The command will return the *patch ID* of the newly created patch. It does **not** actually create a release or upload anything.

```powershell
tp patch {tag} {dir}
tp patch latest path_to_app/
```
Create an incremental patch with file differences included in the patch. The command will return the *patch ID* of the newly created patch. It does **not** actually create a release or upload anything.

```powershell
tp commit {tag} {patch_guid}
tp commit latest base:271df855-2056-4bd5-b6ba-f7d14857820e
```
Upload the patch and make this the newest release for the given tag.

```powershell
tp restore {tag} {dir}
tp restore latest path_to_app/
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