import { apiGet, apiPost } from "@/services/api/request";

export const AUTH_TOKEN_KEY = "infinite-canvas-auth-token-v1";

export type UserRole = "guest" | "user" | "admin";

export type AuthUser = {
    id: string;
    username: string;
    displayName: string;
    avatarUrl: string;
    role: UserRole;
    credits: number;
    createdAt: string;
    updatedAt: string;
};

export type AuthSession = {
    token: string;
    user: AuthUser;
    emailVerificationRequired?: boolean;
    verificationEmailSent?: boolean;
};

export type AuthPayload = {
    username: string;
    password: string;
    email?: string;
};

export async function login(payload: AuthPayload) {
    return apiPost<AuthSession>("/api/auth/login", payload);
}

export async function register(payload: AuthPayload) {
    return apiPost<AuthSession>("/api/auth/register", payload);
}

export async function fetchCurrentUser(token?: string) {
    return apiGet<AuthUser>("/api/auth/me", undefined, token);
}

export async function resendVerificationEmail(email: string) {
    return apiPost<{ emailVerificationRequired: boolean; verificationEmailSent: boolean }>("/api/auth/verification/resend", { email });
}

export async function requestPasswordReset(email: string) {
    return apiPost<boolean>("/api/auth/forgot-password", { email });
}

export async function verifyEmail(token: string) {
    return apiPost<boolean>("/api/auth/verify-email", { token });
}

export async function resetPassword(token: string, password: string) {
    return apiPost<AuthSession>("/api/auth/reset-password", { token, password });
}
