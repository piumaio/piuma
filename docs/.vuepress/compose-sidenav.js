import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __filename = fileURLToPath(import.meta.url)

const sidebar = {
    baseOption: undefined,

    getSidebar(HomeTitle = "Home") {
        const root = getRoot() + '/docs'
        const dirs = fs.readdirSync(root, { withFileTypes: true }).filter(f => f.name !== '.vuepress' && f.isDirectory())
        return [{
            text: HomeTitle,
            link: !!this.baseOption ? this.baseOption : '/',
            collapsible: false,
            children: []
        }, ...getSidebarItems(dirs, root, !!this.baseOption ? this.baseOption : '')]
    }
}

function getRootDir() {
    return path.resolve(process.cwd())
}

function collapseSameName(dirName, childrens) {
    const childI = childrens.findIndex(c => `${c.link || c}`.toUpperCase().endsWith(`/${dirName.toUpperCase()}/`))
    if (childI !== -1) {
        const child = childrens.splice(childI, 1)[0]
        return child.link || child
    }
}

function getSidebarItems(dirs, root, parentPath = "") {
    return dirs.reduce((stack, dir) => {
        const childs = fs.readdirSync(path.resolve(root, dir.name), { withFileTypes: true })
        const childsFolders = childs.filter(d => d.isDirectory())
        const hasReadme = !!childs.find(d => d.name == "README.md")
        if (childsFolders.length) {
            const children = getSidebarItems(childsFolders, `${root}/${dir.name}`, `${parentPath}/${dir.name}`)
            stack.push({
                text: dir.name.charAt(0).toUpperCase() + dir.name.slice(1),
                link: hasReadme ? `${parentPath}/${dir.name}/` : collapseSameName(dir.name, children),
                collapsible: !!children.length,
                children
            })
        } else if (hasReadme) {
            stack.push({
                text: dir.name.charAt(0).toUpperCase() + dir.name.slice(1),
                link: `${parentPath}/${dir.name}/`,
                collapsible: false,
                children: []
            })
        }
        return stack.sort((a, b) => (a.link || a) > (b.link || b) ? 1 : -1) // sort by name
            .sort((a, b) => !!(a.link || typeof a == "string") + !!a.collapsible <= !!(b.link || typeof b == "string") + !!b.collapsible ? 1 : -1) // sort by collapse and url path
            .sort((a, b) => a.text.includes("QuickStart") ? -1 : 1) // sort by text
    }, [])
}

function getRoot() {
    tryFindBase()
    let root

    if (!!sidebar.baseOption) {
        root = path.join(getRootDir(), sidebar.baseOption)
    } else {
        root = getRootDir()
    }
    return root
}

function tryFindBase() {
    try {
        let config = path.join(getRootDir(), '/docs/.vuepress/config.js')
        let contents = fs.readFileSync(config, 'utf8')
        let base = contents.match(/(?<="?base"?:+\s?").+(?=")/)[0]
        sidebar.baseOption = base
    } catch (err) {
        console.log("auto-sidebar: Base option not found.")
    }
}

export const getSidebar = () => sidebar.getSidebar()