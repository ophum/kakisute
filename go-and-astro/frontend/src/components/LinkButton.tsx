import { buttonVariants } from "./ui/button";

export function LinkButton({ href, children }: { href: string; children: React.ReactNode }) {
    return <a href={href} className={buttonVariants({ variant: "link", size: "sm" })}>{children}</a>;
}
