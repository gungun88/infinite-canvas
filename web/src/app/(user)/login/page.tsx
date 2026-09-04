"use client";

import { LockOutlined, MailOutlined, UserOutlined } from "@ant-design/icons";
import { App, Button, Form, Input, Modal, Segmented, Space } from "antd";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useRef, useState } from "react";

import {
    fetchCurrentUser,
    requestPasswordReset,
    resetPassword,
    resendVerificationEmail,
    verifyEmail,
} from "@/services/api/auth";
import { useConfigStore } from "@/stores/use-config-store";
import { useUserStore } from "@/stores/use-user-store";

type LoginFormValues = {
    username: string;
    email?: string;
    password: string;
    confirmPassword?: string;
};

type RecoveryFormValues = {
    email: string;
};

type ResetFormValues = {
    password: string;
    confirmPassword: string;
};

function safeRedirect(value: string | null): string {
    const cleaned = (value ?? "").replace(/[\t\n\r]/g, "");
    if (!cleaned.startsWith("/") || cleaned.startsWith("//") || cleaned.startsWith("/\\")) return "/";
    return cleaned;
}

function cleanLoginURL(redirect: string) {
    window.history.replaceState(null, "", `/login${redirect === "/" ? "" : `?redirect=${encodeURIComponent(redirect)}`}`);
}

export default function LoginPage() {
    return (
        <Suspense fallback={null}>
            <LoginContent />
        </Suspense>
    );
}

function LoginContent() {
    const { message } = App.useApp();
    const router = useRouter();
    const searchParams = useSearchParams();
    const login = useUserStore((state) => state.login);
    const register = useUserStore((state) => state.register);
    const setSession = useUserStore((state) => state.setSession);
    const isLoading = useUserStore((state) => state.isLoading);
    const allowRegister = useConfigStore((state) => state.publicSettings?.auth?.allowRegister !== false);
    const [mode, setMode] = useState<"login" | "register">("login");
    const [recoveryMode, setRecoveryMode] = useState<"reset" | "verify">("reset");
    const [recoveryOpen, setRecoveryOpen] = useState(false);
    const [recoveryLoading, setRecoveryLoading] = useState(false);
    const [verificationEmail, setVerificationEmail] = useState("");
    const [resetToken, setResetToken] = useState("");
    const [resetLoading, setResetLoading] = useState(false);
    const [form] = Form.useForm<LoginFormValues>();
    const [recoveryForm] = Form.useForm<RecoveryFormValues>();
    const [resetForm] = Form.useForm<ResetFormValues>();
    const handledToken = useRef("");
    const handledVerifyToken = useRef("");
    const handledResetToken = useRef("");
    const redirect = safeRedirect(searchParams.get("redirect"));

    useEffect(() => {
        const token = searchParams.get("token");
        const error = searchParams.get("error");
        if (error) message.error(formatAuthError(error));
        if (!token || handledToken.current === token) return;
        handledToken.current = token;
        cleanLoginURL(redirect);
        void fetchCurrentUser(token)
            .then((user) => {
                setSession(token, user);
                message.success("登录成功");
                router.replace(redirect);
                router.refresh();
            })
            .catch(() => message.error("登录状态已失效，请重新登录"));
    }, [message, redirect, router, searchParams, setSession]);

    useEffect(() => {
        const verifyToken = searchParams.get("verify_token");
        if (verifyToken && handledVerifyToken.current !== verifyToken) {
            handledVerifyToken.current = verifyToken;
            window.history.replaceState(null, "", `/login${redirect === "/" ? "" : `?redirect=${encodeURIComponent(redirect)}`}`);
            void verifyEmail(verifyToken)
                .then(() => message.success("邮箱验证成功，请重新登录"))
                .catch((error) => message.error(formatAuthError(error instanceof Error ? error.message : "")));
            return;
        }
        const nextResetToken = searchParams.get("reset_token");
        if (nextResetToken && handledResetToken.current !== nextResetToken) {
            handledResetToken.current = nextResetToken;
            setResetToken(nextResetToken);
            window.history.replaceState(null, "", `/login${redirect === "/" ? "" : `?redirect=${encodeURIComponent(redirect)}`}`);
        }
    }, [message, redirect, searchParams]);

    useEffect(() => {
        if (!allowRegister && mode === "register") setMode("login");
    }, [allowRegister, mode]);

    const submit = async (values: LoginFormValues) => {
        try {
            if (mode === "register") {
                if (!allowRegister) {
                    message.error("当前未开放注册");
                    return;
                }
                if (values.password !== values.confirmPassword) {
                    message.error("两次输入的密码不一致");
                    return;
                }
                const session = await register({ username: values.username, password: values.password, email: values.email });
                if (session.emailVerificationRequired) {
                    setVerificationEmail(values.email || "");
                    setMode("login");
                    form.resetFields(["password", "confirmPassword"]);
                    form.setFieldsValue({ username: values.username });
                    message.success(session.verificationEmailSent ? "验证邮件已发送，请查收邮箱" : "账号已创建，请稍后重新发送验证邮件");
                    return;
                }
                message.success("注册成功");
                router.replace(redirect);
                router.refresh();
                return;
            }

            const user = await login({ username: values.username, password: values.password });
            message.success("登录成功");
            router.replace(user.role === "admin" ? redirect : "/");
            router.refresh();
        } catch (error) {
            message.error(formatAuthError(error instanceof Error ? error.message : ""));
        }
    };

    const submitRecovery = async (values: RecoveryFormValues) => {
        setRecoveryLoading(true);
        try {
            if (recoveryMode === "reset") {
                await requestPasswordReset(values.email);
                message.success("如果邮箱存在，重置链接已发送");
            } else {
                const result = await resendVerificationEmail(values.email);
                message.success(result.verificationEmailSent ? "验证邮件已发送" : "验证邮件暂未发送，请稍后再试");
            }
            setRecoveryOpen(false);
            recoveryForm.resetFields();
        } catch (error) {
            message.error(formatAuthError(error instanceof Error ? error.message : ""));
        } finally {
            setRecoveryLoading(false);
        }
    };

    const submitReset = async (values: ResetFormValues) => {
        if (values.password !== values.confirmPassword) {
            message.error("两次输入的密码不一致");
            return;
        }
        setResetLoading(true);
        try {
            const session = await resetPassword(resetToken, values.password);
            setSession(session.token, session.user);
            message.success("密码已重置");
            router.replace(redirect);
            router.refresh();
        } catch (error) {
            message.error(formatAuthError(error instanceof Error ? error.message : ""));
        } finally {
            setResetLoading(false);
        }
    };

    const openRecovery = (nextMode: "reset" | "verify") => {
        setRecoveryMode(nextMode);
        setRecoveryOpen(true);
        recoveryForm.setFieldsValue({ email: verificationEmail || form.getFieldValue("email") || "" });
    };

    if (resetToken) {
        return (
            <main className="flex h-full min-h-0 items-center justify-center overflow-y-auto bg-background px-6 py-10">
                <section className="w-full max-w-[420px]">
                    <div className="mb-7 text-center">
                        <Logo />
                        <h1 className="text-3xl font-semibold tracking-normal text-stone-950 dark:text-stone-100">重置密码</h1>
                    </div>
                    <Form<ResetFormValues> form={resetForm} layout="vertical" requiredMark={false} onFinish={submitReset}>
                        <Form.Item name="password" label="新密码" rules={[{ required: true, min: 8, message: "密码至少 8 位" }]}>
                            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
                        </Form.Item>
                        <Form.Item name="confirmPassword" label="确认新密码" rules={[{ required: true, message: "请再次输入新密码" }]}>
                            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
                        </Form.Item>
                        <Button block type="primary" htmlType="submit" loading={resetLoading}>
                            提交并登录
                        </Button>
                    </Form>
                </section>
            </main>
        );
    }

    return (
        <main className="flex h-full min-h-0 items-center justify-center overflow-y-auto bg-background bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] px-6 py-10 [background-size:16px_16px] dark:bg-[radial-gradient(rgba(245,245,244,.16)_1px,transparent_1px)]">
            <section className="w-full max-w-[420px]">
                <div className="mb-7 text-center">
                    <Logo />
                    <h1 className="text-3xl font-semibold tracking-normal text-stone-950 dark:text-stone-100">{mode === "register" ? "创建账号" : "账号登录"}</h1>
                </div>

                <Form<LoginFormValues> form={form} layout="vertical" size="large" requiredMark={false} onFinish={submit}>
                    <Form.Item>
                        <Segmented
                            block
                            value={mode}
                            onChange={(value) => setMode(value as "login" | "register")}
                            options={allowRegister ? [{ label: "登录", value: "login" }, { label: "注册", value: "register" }] : [{ label: "登录", value: "login" }]}
                        />
                    </Form.Item>
                    <Form.Item name="username" label="用户名" rules={[{ required: true, message: "请输入用户名" }]}>
                        <Input prefix={<UserOutlined />} autoComplete="username" />
                    </Form.Item>
                    {mode === "register" ? (
                        <Form.Item name="email" label="邮箱" rules={[{ required: true, message: "请输入邮箱" }, { type: "email", message: "邮箱格式不正确" }]}>
                            <Input prefix={<MailOutlined />} autoComplete="email" />
                        </Form.Item>
                    ) : null}
                    <Form.Item name="password" label="密码" rules={[{ required: true, message: "请输入密码" }]}>
                        <Input.Password prefix={<LockOutlined />} autoComplete={mode === "register" ? "new-password" : "current-password"} />
                    </Form.Item>
                    {mode === "register" ? (
                        <Form.Item name="confirmPassword" label="确认密码" rules={[{ required: true, message: "请再次输入密码" }]}>
                            <Input.Password prefix={<LockOutlined />} autoComplete="new-password" />
                        </Form.Item>
                    ) : null}
                    <Space orientation="vertical" size={12} style={{ width: "100%" }}>
                        <Button block type="primary" htmlType="submit" loading={isLoading}>
                            {mode === "register" ? "注册" : "登录"}
                        </Button>
                    </Space>
                    <div className="mt-3 flex items-center justify-between text-xs">
                        <Button type="link" className="!px-0 !text-xs" onClick={() => openRecovery("reset")}>
                            忘记密码
                        </Button>
                        {mode === "login" ? (
                            <Button type="link" className="!px-0 !text-xs" onClick={() => openRecovery("verify")}>
                                重新发送验证邮件
                            </Button>
                        ) : null}
                    </div>
                </Form>

                <div className="space-y-3 pt-6">
                    <div className="flex items-center gap-3">
                        <div className="h-px flex-1 bg-stone-200 dark:bg-stone-800" />
                        <span className="text-xs text-stone-500 dark:text-stone-400">或者使用其他登录方式</span>
                        <div className="h-px flex-1 bg-stone-200 dark:bg-stone-800" />
                    </div>
                    <div className="flex flex-col gap-3">
                        <Button block size="large" className="!h-12 !justify-center !gap-2" href={`/api/auth/google/authorize?redirect=${encodeURIComponent(redirect)}`}>
                            <GoogleIcon />
                            <span className="font-medium">使用 Google 登录</span>
                        </Button>
                        <Button block size="large" className="!h-12 !justify-center !gap-2" href={`/api/auth/doingfb/authorize?redirect=${encodeURIComponent(redirect)}`}>
                            <img alt="" className="h-5 w-5 rounded-full object-contain" src="https://img.doingfb.com/%E9%AB%98%E6%B8%85logo%E8%BF%98%E5%8E%9F.png" />
                            <span className="font-medium">使用 DoingFB 登录</span>
                        </Button>
                    </div>
                </div>
            </section>

            <Modal
                open={recoveryOpen}
                title={recoveryMode === "reset" ? "找回密码" : "重新发送验证邮件"}
                onCancel={() => setRecoveryOpen(false)}
                footer={null}
                destroyOnClose
                centered
            >
                <Form form={recoveryForm} layout="vertical" requiredMark={false} onFinish={submitRecovery}>
                    <Form.Item name="email" label="邮箱" rules={[{ required: true, message: "请输入邮箱" }, { type: "email", message: "邮箱格式不正确" }]}>
                        <Input size="large" prefix={<MailOutlined />} placeholder="you@example.com" autoComplete="email" />
                    </Form.Item>
                    <Button type="primary" htmlType="submit" block size="large" loading={recoveryLoading}>
                        {recoveryMode === "reset" ? "发送重置邮件" : "发送验证邮件"}
                    </Button>
                </Form>
            </Modal>
        </main>
    );
}

function Logo() {
    return (
        <span
            className="mx-auto mb-4 block size-12 bg-stone-950 dark:bg-stone-100"
            style={{
                mask: "url(/logo.svg) center / contain no-repeat",
                WebkitMask: "url(/logo.svg) center / contain no-repeat",
            }}
            aria-label="无限画布"
        />
    );
}

function formatAuthError(error: string) {
    const messages: Record<string, string> = {
        email_not_verified: "邮箱尚未验证，请先查收验证邮件",
        email_exists: "该邮箱已注册",
        email_required: "请输入邮箱",
        email_not_configured: "邮件服务尚未配置",
        invalid_email: "邮箱格式不正确",
        registration_disabled: "当前未开放注册",
        invalid_reset_token: "重置链接已失效，请重新申请",
        invalid_verification_token: "验证链接已失效，请重新发送验证邮件",
        google_oauth_not_configured: "Google 登录尚未配置",
        doingfb_oauth_not_configured: "DoingFB 登录尚未配置",
        google_oauth_state_invalid: "Google 授权状态已失效，请重新登录",
        doingfb_oauth_state_invalid: "DoingFB 授权状态已失效，请重新登录",
        google_token_failed: "Google 授权失败，请重试",
        doingfb_token_failed: "DoingFB 授权失败，请重试",
    };
    return messages[error] || error || "操作失败";
}

function GoogleIcon() {
    return (
        <svg viewBox="0 0 24 24" aria-hidden="true" className="h-5 w-5">
            <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" />
            <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98-.66-2.23-1.06-3.71-1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
            <path fill="#FBBC05" d="M5.84 14.1A6.61 6.61 0 0 1 5.5 12c0-.73.12-1.43.34-2.1V7.06H2.18A10.96 10.96 0 0 0 1 12c0 1.77.42 3.45 1.18 4.94l3.66-2.84z" />
            <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06L5.84 9.9C6.71 7.31 9.14 5.38 12 5.38z" />
        </svg>
    );
}
