const TabWalletSend = document.querySelector("#TabContentWalletSend");
const TabContentEvil = document.querySelector("#TabContentEvil");

// TabWalletSend.addEventListener("change", (e) => {
//   const trValInp = e.target.matches(".rc-wallet-tr-input");
//   if (trValInp) {
//     const inp = TabWalletSend.querySelector(".rc-wallet-tr-input");
//     console.log(inp);
//   }
// });

TabWalletSend.addEventListener("animationend", (e) => RcFadeOutRemove(e));
TabContentEvil.addEventListener("animationend", (e) => RcFadeOutRemove(e));

function RcFadeOutRemove(e) {
  const isHideElement = e.target.matches(".rc-fade-out");
  if (isHideElement) {
    e.target.style.display = "none";
  }
}
