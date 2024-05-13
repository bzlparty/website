# buildifier: disable=module-docstring
def _impl(ctx):
    for m in ctx.modules:
        for t in m.tags.repository:
            docs_repository(
                name = t.name,
                commit = t.commit,
                project = t.project,
            )

def _docs_repo_impl(ctx):
    downloader = ctx.path(ctx.attr._downloader)
    output = ctx.execute([downloader, ctx.attr.project, ctx.attr.commit])
    targets = []
    for f in output.stdout.split("\n"):
        path = "/".join(f.split("/")[1:-1])
        if path == "":
            continue
        target = "//%s:files" % path
        if target in targets:
            continue
        targets.append(target)
        ctx.file("%s/BUILD" % path, """\
load("@rules_pkg//pkg:mappings.bzl", "pkg_files")

pkg_files(
    name = "files",
    srcs = glob(["*.md"]),
    prefix = "{}",
    visibility = ["//visibility:public"],
)
    """.format(path))

    targets.append(":files")
    ctx.file("commit.info", ctx.attr.commit)
    ctx.file("BUILD.bazel", """\
load("@rules_pkg//pkg:mappings.bzl", "pkg_files", "pkg_filegroup")

pkg_files(
    name = "files",
    srcs = glob(["*.md"]) + ["commit.info"],
    visibility = ["//visibility:public"],
)

pkg_filegroup(
    name = "docs",
    srcs = {targets},
    prefix = "{project}",
    visibility = ["//visibility:public"],
)
    """.format(targets = targets, project = ctx.attr.project))
    ctx.file("WORKSPACE", "workspace(name = \"{name}\")".format(name = ctx.name))

docs_repository = repository_rule(
    _docs_repo_impl,
    attrs = {
        "_downloader": attr.label(default = ":docs_download.sh"),
        "commit": attr.string(),
        "project": attr.string(),
    },
)

docs = module_extension(
    _impl,
    tag_classes = {
        "repository": tag_class(attrs = {
            "name": attr.string(),
            "commit": attr.string(),
            "project": attr.string(),
        }),
    },
)

def _css_impl(ctx):
    ctx.download(
        output = "lissom.css",
        url = "https://raw.githubusercontent.com/lissomware/css/00a08324134616a60bc3f21ebda92b349d19b731/lissom.min.css",
        integrity = "sha384-9+IaBL1uGqkYxp0d/br3fzTKrtgMZ0o4H5YyzvYnPglnCyMAqVDpqOPaVckHXlKQ",
    )
    ctx.file("BUILD", """\
package(default_visibility = ["//visibility:public"])
exports_files(["lissom.css"])
alias(
  name = "lissom",
  actual = "lissom.css",
)
""")

css_lib = repository_rule(_css_impl)

css = module_extension(lambda _: css_lib(name = "css"))
