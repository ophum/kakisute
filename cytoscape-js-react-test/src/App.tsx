import { useEffect, useRef, useState } from "react";
import cytoscape from "cytoscape";
import yaml from "js-yaml";

type ElemType = (ElemNode | ElemEdge)[];

type ElemNode = {
  data: {
    id: string;
    label?: string;
    parent?: string;
  };
  position?: {
    x: number;
    y: number;
  };
};

type ElemEdge = {
  data: {
    id: string;
    source: string;
    target: string;
    label?: string;
    style?: string;
  };
};

interface Manifest {
  hosts: Host[];
}
interface Host {
  name: string;
  processes: Process[];
}
interface Process {
  name: string;
  listeners: Listener[];
  connections: Connection[];
}
interface Listener {
  port: number;
  protocol: string;
}
interface Connection {
  host: string;
  port: number;
  protocol: string;
}
function manifestToElems(manifest: string): ElemType {
  if (manifest === "") return [];
  const m = yaml.load(manifest) as Manifest;
  console.log(m);
  const ret: ElemType = [];
  const listeners: { [key: string]: { [port: number]: string } } = {};
  for (const host of m.hosts) {
    for (const process of host.processes) {
      for (const listener of process.listeners) {
        if (listeners[host.name] === undefined) {
          listeners[host.name] = {};
        }
        listeners[host.name][
          listener.port
        ] = `${host.name}-listen-${listener.protocol}-${listener.port}`;
      }
    }
  }
  ret.push({
    data: {
      id: "listener",
      label: "listener",
    },
  });
  for (const host of m.hosts) {
    const h = {
      data: {
        id: `${host.name}-listen`,
        label: `${host.name}`,
        parent: "listener",
      },
    };
    ret.push(h);
    for (const process of host.processes) {
      if (process.listeners.length === 0) {
        continue;
      }
      const p = {
        data: {
          id: `${host.name}-listen-${process.name}`,
          label: `${process.name}`,
          parent: h.data.id,
        },
      };
      ret.push(p);
      for (const listener of process.listeners) {
        ret.push({
          data: {
            id: listeners[host.name][listener.port],
            label: `:${listener.port}`,
            parent: p.data.id,
          },
        });
      }
    }
  }

  for (const host of m.hosts) {
    const h = {
      data: {
        id: `${host.name}`,
        label: `${host.name}`,
      },
    };
    ret.push(h);
    for (const process of host.processes) {
      const p = {
        data: {
          id: `${host.name}-${process.name}`,
          label: `${process.name}`,
          parent: h.data.id,
        },
      };
      ret.push(p);
      for (const connection of process.connections) {
        console.log({
          host: host.name,
          process: process.name,
          targethost: connection.host,
          targetport: connection.port,

          source: p.data.id,
          target: listeners[connection.host][connection.port],
        });
        ret.push({
          data: {
            id: `${host.name}-${process.name}-${connection.host}-${connection.port}-edge`,
            source: p.data.id,
            target: listeners[connection.host][connection.port],
            label: listeners[connection.host][connection.port],
          },
        });
      }
    }
  }
  return ret;
}

function App() {
  const cytoEl = useRef<HTMLDivElement>(null);
  const [manifest, setManifest] = useState("");

  useEffect(() => {
    let elems: ElemType = [];
    try {
      elems = manifestToElems(manifest);
    } catch (e: unknown) {
      console.log(e);
      return;
    }

    const cyto = cytoscape({
      container: cytoEl.current,
      elements: elems,
      style: [
        {
          selector: "node",
          style: {
            width: "40px",
            height: "20px",
            label: "data(label)",
            "border-width": 2,
            "text-valign": "bottom",
            shape: "rectangle",
          } as const,
        },
        {
          selector: "edge",
          style: {
            //"curve-style": "taxi",
            "curve-style": "bezier",
            //"curve-style": "round-segments",
            "target-arrow-shape": "triangle",
            "line-style": "data(style)",
          },
        },
        {
          selector: ":parent",
          style: {
            "text-valign": "bottom",
            "text-halign": "center",
            padding: "10px",
          },
        },
      ],
      layout: {
        name: "circle",
        padding: 5,
      },
    });
    return () => {
      cyto.destroy();
    };
  }, [manifest]);
  return (
    <div
      style={{
        display: "flex",
        flexDirection: "row",
      }}
    >
      <textarea
        value={manifest}
        onChange={(e) => {
          setManifest(e.target.value);
        }}
        style={{
          width: "20vw",
          height: "100vh",
        }}
      ></textarea>
      <div
        ref={cytoEl}
        style={{
          width: "80vw",
          height: "100vh",
        }}
      />
    </div>
  );
}

export default App;
