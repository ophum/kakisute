import { useState, useEffect } from "react";
import { Button } from "./ui/button";

export function UserInfo() {
    const [info, setInfo] = useState(null);

    const load = async() => {
        setInfo(null);
        const res = await fetch("/api/user");
        const data = await res.json();
        setInfo(data);
    }
    useEffect(() => {
        load();
    }, []);
    return (
        <div>
            <Button onClick={load}>Reload</Button>
            {info ? (
                <pre>{JSON.stringify(info, null, 2)}</pre>
            ) : (
                <p>No user info available.</p>
            )}
        </div>
    );
}
