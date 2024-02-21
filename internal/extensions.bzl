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
    ctx.file("BUILD.bazel", """\
load("@rules_pkg//pkg:mappings.bzl", "pkg_files", "pkg_filegroup")

pkg_files(
    name = "files",
    srcs = glob(["*.md"]),
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
