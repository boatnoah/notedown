interface Operation {
  clientID: string;
  value: string;
  charID: string;
  action: "INSERT" | "DELETE";
  position: number;
}

export default Operation;
