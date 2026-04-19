import { useState } from "react";
import { useAuth } from "../context/AuthContext";
import { useNavigate } from "react-router-dom";
import { Shield } from "lucide-react";

export default function Login() {
  const { sendCode, verifyCode } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [step, setStep] = useState<"email" | "code">("email");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSendCode = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await sendCode(email);
      setStep("code");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to send code");
    } finally {
      setLoading(false);
    }
  };

  const handleVerifyCode = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      await verifyCode(email, code);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Invalid code");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-5">
      <div className="w-full max-w-[400px]">
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-12 h-12 bg-(--color-accent) rounded-xl mb-4">
            <Shield size={24} color="white" />
          </div>
          <h1 className="text-2xl font-bold mb-2">StackTrace</h1>
          <p className="text-(--color-text-secondary) text-sm">
            {step === "email"
              ? "Enter your email to sign in"
              : "Enter the 6-digit code sent to your email"}
          </p>
        </div>

        <div className="card">
          {step === "email" ? (
            <form onSubmit={handleSendCode}>
              <label className="form-label">Email address</label>
              <input
                className="input mb-4"
                type="email"
                placeholder="you@company.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoFocus
              />
              <button
                className="btn btn-primary w-full"
                type="submit"
                disabled={loading || !email}
              >
                {loading ? "Sending..." : "Send verification code"}
              </button>
            </form>
          ) : (
            <form onSubmit={handleVerifyCode}>
              <label className="form-label">Verification code</label>
              <input
                className="input mb-4 text-2xl text-center tracking-[8px] font-semibold"
                type="text"
                placeholder="000000"
                value={code}
                onChange={(e) =>
                  setCode(e.target.value.replace(/\D/g, "").slice(0, 6))
                }
                required
                autoFocus
                maxLength={6}
              />
              <button
                className="btn btn-primary w-full mb-3"
                type="submit"
                disabled={loading || code.length !== 6}
              >
                {loading ? "Verifying..." : "Verify and sign in"}
              </button>
              <button
                type="button"
                className="btn btn-ghost w-full"
                onClick={() => {
                  setStep("email");
                  setCode("");
                  setError("");
                }}
              >
                Use a different email
              </button>
            </form>
          )}

          {error && (
            <p className="text-(--color-danger) text-[13px] mt-3 text-center">
              {error}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}
