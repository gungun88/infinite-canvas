import type { LucideIcon } from "lucide-react";
import { FileText, ImagePlus, Images, Layers3, Maximize2, Video, WandSparkles } from "lucide-react";

export type NavigationTool = {
    label: string;
    icon: LucideIcon;
    slug?: string;
    href?: string;
    activePrefix?: string;
};

export const navigationTools = [
    {
        slug: "canvas",
        label: "我的画布",
        icon: Maximize2,
    },
    {
        slug: "image",
        label: "生图工作台",
        icon: ImagePlus,
    },
    {
        slug: "video",
        label: "视频创作台",
        icon: Video,
    },
    {
        slug: "workflows",
        label: "工作流",
        icon: WandSparkles,
    },
    {
        href: "/director/index.html",
        activePrefix: "/director",
        label: "导演台",
        icon: Layers3,
    },
    {
        slug: "prompts",
        label: "提示词库",
        icon: FileText,
    },
    {
        slug: "assets",
        label: "我的素材",
        icon: Images,
    },
] as const satisfies readonly NavigationTool[];

export type NavigationToolSlug = Extract<(typeof navigationTools)[number], { slug: string }>["slug"];
