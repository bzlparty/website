# buildifier: disable=module-docstring
load("@rules_pkg//pkg:providers.bzl", "PackageFilegroupInfo", "PackageFilesInfo")
load("@rules_pkg//pkg:tar.bzl", "pkg_tar")

# buildifier: disable=function-docstring
def _render_html_impl(ctx):
    outputs = []
    package_info = struct(
        name = ctx.attr.package,
        files = [],
        template = ctx.file.template.path,
    )
    pkg_files = []
    for (data, _) in ctx.attr.src[PackageFilegroupInfo].pkg_files:
        dest_src_map = {}
        for (s, f) in data.dest_src_map.items():
            target = s.replace("README", "index").replace(".md", ".html")

            # print(target)
            # target = "/".join(target.split("/")[1:])
            # print(target)
            output = ctx.actions.declare_file(target)
            dest_src_map[target] = output
            outputs.append(output)
            package_info.files.append({
                "source_path": f.path,
                "target": target,
                "target_path": output.path,
            })
        pkg_files.append((PackageFilesInfo(dest_src_map = dest_src_map), ctx.attr.package))

    args = ctx.actions.args()
    args.add("-jsonConfig", json.encode(package_info))

    ctx.actions.run(
        inputs = ctx.files.src + [ctx.file.template],
        outputs = outputs,
        arguments = [args],
        mnemonic = "HtmlRender",
        progress_message = "Render: %s" % ctx.attr.name,
        executable = ctx.executable.renderer,
    )

    runfiles = ctx.runfiles(files = ctx.files.src + [ctx.file.template])

    return [
        DefaultInfo(files = depset(outputs), runfiles = runfiles),
        PackageFilegroupInfo(pkg_files = pkg_files),
    ]

render_html = rule(
    _render_html_impl,
    attrs = {
        "src": attr.label(
            providers = [PackageFilegroupInfo],
        ),
        "package": attr.string(),
        "template": attr.label(allow_single_file = True, default = ":package.html"),
        "renderer": attr.label(
            default = ":render_html",
            cfg = "exec",
            executable = True,
        ),
    },
)

# buildifier: disable=function-docstring
def html_pkg_tar(name, srcs, **kwargs):
    outputs = []
    for (package, src) in srcs.items():
        output = "%s_html" % package
        render_html(
            name = output,
            src = src,
            package = package,
        )
        outputs.append(output)
    pkg_tar(
        name = name,
        srcs = outputs,
        **kwargs
    )
