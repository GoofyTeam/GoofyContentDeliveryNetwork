import * as React from "react";
import { AuthState, AuthFunctions, User } from "./hooksTypes";

interface AuthContext extends AuthState, AuthFunctions {}

const AuthContext = React.createContext<AuthContext | null>(null);

const USER_KEY = "goofycontentdeliverynetwork.auth.user";
const USER_TOKEN = "goofycontentdeliverynetwork.auth.token";

function getStoredUser() {
  const user = localStorage.getItem(USER_KEY);

  if (user) {
    return JSON.parse(user);
  }

  console.warn("no user found");

  return null;
}

function getStoredJwt() {
  return localStorage.getItem(USER_TOKEN);
}

function setStoredUser(user: User | null) {
  if (user) {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  } else {
    localStorage.removeItem(USER_KEY);
  }
}

function updateStoredUser(user: Partial<User>) {
  const storedUser = getStoredUser();

  if (storedUser) {
    setStoredUser({ ...storedUser, ...user });
  }
}

function setStoredJwt(jwt: string | null) {
  if (jwt) {
    localStorage.setItem(USER_TOKEN, jwt);
  } else {
    localStorage.removeItem(USER_TOKEN);
  }
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = React.useState<User | null>(getStoredUser());
  const [accessToken, setAccessToken] = React.useState<string | null>(
    getStoredJwt()
  );

  const isAuth = !!user;

  const logout = React.useCallback(async () => {
    setStoredUser(null);
    setStoredJwt(null);
    setUser(null);
    setAccessToken(null);
  }, []);

  const login = React.useCallback(async (user: User, jwt: string) => {
    setStoredUser(user);
    setStoredJwt(jwt);
    setUser(user);
    setAccessToken(jwt);
  }, []);

  const update = React.useCallback(async (user: Partial<User>) => {
    updateStoredUser(user);
    setUser((prev) => {
      if (!prev) {
        return null;
      }

      return {
        ...prev,
        ...user,
      };
    });
  }, []);

  React.useEffect(() => {
    setUser(getStoredUser());
  }, []);

  return (
    <AuthContext.Provider
      value={{ isAuth, user, accessToken, login, logout, update }}
    >
      {children}
    </AuthContext.Provider>
  );
}

function useAuth() {
  const context = React.useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export default useAuth;
